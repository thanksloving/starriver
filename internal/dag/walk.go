package dag

import (
	"fmt"
	"sync"

	"github.com/thanksloving/starriver"
	"github.com/thanksloving/starriver/helper"
)

// Walker is used to walk every vertex of a graph in parallel.
//
// A vertex will only be walked when the dependencies of that vertex have
// been walked. If two vertices can be walked at the same time, they will be.
//
// Update can be called to update the graph. This can be called even during
// a walk, changing vertices/edges mid-walk. This should be done carefully.
// If a vertex is removed but has already been executed, the result of that
// execution (any error) is still returned by Wait. Changing or re-adding
// a vertex that has already executed has no effect. Changing edges of
// a vertex that has already executed has no effect.
//
// Non-parallelism can be enforced by introducing a lock in your callback
// function. However, the goroutine overhead of a walk will remain.
// Walker will create V*2 goroutines (one for each vertex, and dependency
// waiter for each vertex). In general this should be of no concern unless
// there are a huge number of vertices.
//
// The walk is depth first by default. This can be changed with the Reverse
// option.
//
// A single walker is only valid for one graph walk. After the walk is complete
// you must construct a new walker to walk again. State for the walk is never
// deleted in case vertices or edges are changed.
type Walker struct {
	// DataContext Actually is an instance of starriver.DataContext
	DataContext starriver.DataContext

	// Callback is what is called for each vertex
	Callback WalkFunc

	// Reverse, if true, causes the source of an edge to depend on a target.
	// When false (default), the target depends on the source.
	Reverse bool

	// changeLock must be held to modify any of the fields below. Only Update
	// should modify these fields. Modifying them outside of Update can cause
	// serious problems.
	changeLock sync.Mutex
	vertices   Set
	edges      Set
	vertexMap  map[Vertex]*walkerVertex

	// wait is done when all vertices have executed. It may become "undone"
	// if new vertices are added.
	wait sync.WaitGroup

	// respMap contains the diagnostics recorded so far for execution,
	// and upstreamFailed contains all the vertices whose problems were
	// caused by upstream failures, and thus whose diagnostics should be
	// excluded from the final set.
	//
	// Readers and writers of either map must hold respLock.
	respMap        map[string]starriver.Response
	upstreamFailed map[string]struct{}
	respLock       sync.Mutex
}

func (w *Walker) init() {
	if w.vertices == nil {
		w.vertices = make(Set)
	}
	if w.edges == nil {
		w.edges = make(Set)
	}
}

type walkerVertex struct {
	// These should only be set once on initialization and never written again.
	// They are not protected by a lock since they don't need to be since
	// they are write-once.

	// DoneCh is closed when this vertex has completed execution, regardless
	// of success.
	//
	// CancelCh is closed when the vertex should cancel execution. If execution
	// is already complete (DoneCh is closed), this has no effect. Otherwise,
	// execution is cancelled as quickly as possible.
	DoneCh   chan struct{}
	CancelCh chan struct{}

	UpEdges []Edge

	// Dependency information. Any changes to any of these fields requires
	// holding DepsLock.
	//
	// DepsCh is sent a single value that denotes whether the upstream deps
	// were successful (no errors). Any value sent means that the upstream
	// dependencies are complete. No other values will ever be sent again.
	//
	// DepsUpdateCh is closed when there is a new DepsCh set.
	DepsCh       chan bool
	DepsUpdateCh chan struct{}
	DepsLock     sync.Mutex

	// Below is not safe to read/write in parallel. This behavior is
	// enforced by changes only happening in Update. Nothing else should
	// ever modify these.
	deps         map[Vertex]chan struct{}
	depsCancelCh chan struct{}
}

// Wait waits for the completion of the walk and returns diagnostics describing
// any problems that arose. Update should be called to populate the walk with
// vertices and edges prior to calling this.
//
// Wait will return as soon as all currently known vertices are complete.
// If you plan on calling Update with more vertices in the future, you
// should not call Wait until after this is done.
func (w *Walker) Wait() starriver.Responses {
	// Wait for completion
	w.wait.Wait()

	var responses = make(starriver.Responses)
	w.respLock.Lock()
	for vertexId, resp := range w.respMap {
		if _, upstreamHasFailed := w.upstreamFailed[vertexId]; upstreamHasFailed {
			continue
		}
		responses[vertexId] = resp
	}
	w.respLock.Unlock()

	return responses
}

func graphObjListToSet(list interface{}) Set {
	result := make(Set)
	switch ls := list.(type) {
	case []Vertex:
		for _, vertex := range ls {
			result.Add(vertex)
		}
	case []Edge:
		for _, edge := range ls {
			result.Add(edge)
		}
	}
	return result
}

// Update updates the currently executing walk with the given graph.
// This will perform a diff of the vertices and edges and update the walker.
// Already completed vertices remain completed (including any errors during
// their execution).
//
// This returns immediately once the walker is updated; it does not wait
// for completion of the walk.
//
// Multiple Updates can be called in parallel. Update can be called at any
// time during a walk.
func (w *Walker) Update(g DAG) {
	w.init()
	var v, e Set
	if g != nil {
		v, e = graphObjListToSet(g.Vertices()), graphObjListToSet(g.Edges())
	} else {
		v = make(Set)
		e = make(Set)
	}

	// Grab the change lock so no more updates happen but also so that
	// no new vertices are executed during this time since we may be
	// removing them.
	w.changeLock.Lock()
	defer w.changeLock.Unlock()

	// Initialize fields
	if w.vertexMap == nil {
		w.vertexMap = make(map[Vertex]*walkerVertex)
	}

	// Calculate all our sets
	newEdges := e.Difference(w.edges)
	oldEdges := w.edges.Difference(e)
	newVerts := v.Difference(w.vertices)
	oldVerts := w.vertices.Difference(v)

	// Add the new vertices
	for _, raw := range newVerts {
		v := raw

		// Add to the waitgroup so our walk is not done until everything finishes
		w.wait.Add(1)

		// Add to our own set so we know about it already
		w.vertices.Add(raw)

		upEdges := make([]Edge, 0)
		for _, ne := range newEdges {
			if ne.(Edge).Target().ID() != v.ID() {
				continue
			}
			upEdges = append(upEdges, ne.(Edge))
		}

		// Initialize the vertex info
		info := &walkerVertex{
			DoneCh:   make(chan struct{}),
			CancelCh: make(chan struct{}),
			deps:     make(map[Vertex]chan struct{}),
			UpEdges:  upEdges,
		}

		// Add it to the map and kick off the walk
		w.vertexMap[v] = info
	}

	// Remove the old vertices
	for _, raw := range oldVerts {
		v := raw

		// Get the vertex info so we can cancel it
		info, ok := w.vertexMap[v]
		if !ok {
			// This vertex for some reason was never in our map. This
			// shouldn't be possible.
			continue
		}

		// Cancel the vertex
		close(info.CancelCh)

		// Delete it out of the map
		delete(w.vertexMap, v)
		w.vertices.Delete(raw)
	}

	// Add the new edges
	changedDeps := make(Set)
	for _, raw := range newEdges {
		edge := raw.(Edge)
		waiter, dep := w.edgeParts(edge)

		// Get the info for the waiter
		waiterInfo, ok := w.vertexMap[waiter]
		if !ok {
			// Vertex doesn't exist... shouldn't be possible but ignore.
			continue
		}

		// Get the info for the dep
		depInfo, ok := w.vertexMap[dep]
		if !ok {
			// Vertex doesn't exist... shouldn't be possible but ignore.
			continue
		}

		// Add the dependency to our waiter
		waiterInfo.deps[dep] = depInfo.DoneCh

		// Record that the deps changed for this waiter
		changedDeps.Add(waiter)
		w.edges.Add(raw)
	}

	// Process removed edges
	for _, raw := range oldEdges {
		edge := raw.(Edge)
		waiter, dep := w.edgeParts(edge)

		// Get the info for the waiter
		waiterInfo, ok := w.vertexMap[waiter]
		if !ok {
			// Vertex doesn't exist... shouldn't be possible but ignore.
			continue
		}

		// Delete the dependency from the waiter
		delete(waiterInfo.deps, dep)

		// Record that the deps changed for this waiter
		changedDeps.Add(waiter)
		w.edges.Delete(raw)
	}

	// For each vertex with changed dependencies, we need to kick off
	// a new waiter and notify the vertex of the changes.
	for _, raw := range changedDeps {
		v := raw
		info, ok := w.vertexMap[v]
		if !ok {
			// Vertex doesn't exist... shouldn't be possible but ignore.
			continue
		}

		// Create a new done channel
		doneCh := make(chan bool, 1)

		// Create the channel we close for cancellation
		cancelCh := make(chan struct{})

		// Build a new deps copy
		deps := make(map[Vertex]<-chan struct{})
		for k, v := range info.deps {
			deps[k] = v
		}

		// Update the update channel
		info.DepsLock.Lock()
		if info.DepsUpdateCh != nil {
			close(info.DepsUpdateCh)
		}
		info.DepsCh = doneCh
		info.DepsUpdateCh = make(chan struct{})
		info.DepsLock.Unlock()

		// Cancel the older waiter
		if info.depsCancelCh != nil {
			close(info.depsCancelCh)
		}
		info.depsCancelCh = cancelCh

		// Start the waiter
		go w.waitDeps(w.DataContext, v, deps, doneCh, cancelCh)
	}

	// Start all the new vertices. We do this at the end so that all
	// the edge waiters and changes are set up above.
	for _, raw := range newVerts {
		v := raw
		go w.walkVertex(v, w.DataContext, w.vertexMap[v])
	}
}

// edgeParts returns the waiter and the dependency, in that order.
// The waiter is waiting on the dependency.
func (w *Walker) edgeParts(e Edge) (Vertex, Vertex) {
	if w.Reverse {
		return e.Source(), e.Target()
	}

	return e.Target(), e.Source()
}

// walkVertex walks a single vertex, waiting for any dependencies before
// executing the callback.
func (w *Walker) walkVertex(v Vertex, dataContext starriver.DataContext, info *walkerVertex) {
	// When we're done executing, lower the waitgroup count
	defer w.wait.Done()

	// When we're done, always close our done channel
	defer close(info.DoneCh)

	// Wait for our dependencies. We create a [closed] deps channel so
	// that we can immediately fall through to load our actual DepsCh.
	var depsSuccess bool
	var depsUpdateCh chan struct{}
	depsCh := make(chan bool, 1)
	depsCh <- true
	close(depsCh)
	for {
		select {
		case <-info.CancelCh:
			// Cancel
			return

		case depsSuccess = <-depsCh:
			// Deps complete! Mark as nil to trigger completion handling.
			depsCh = nil

		case <-depsUpdateCh:
			// New deps, reloop
		}

		// Check if we have updated dependencies. This can happen if the
		// dependencies were satisfied exactly prior to an Update occurring.
		// In that case, we'd like to take into account new dependencies
		// if possible.
		info.DepsLock.Lock()
		if info.DepsCh != nil {
			depsCh = info.DepsCh
			info.DepsCh = nil
		}
		if info.DepsUpdateCh != nil {
			depsUpdateCh = info.DepsUpdateCh
		}
		info.DepsLock.Unlock()

		// If we still have no deps channel set, then we're done!
		if depsCh == nil {
			break
		}
	}

	// If we passed dependencies, we just want to check once more that
	// we're not cancelled, since this can happen just as dependencies pass.
	select {
	case <-info.CancelCh:
		// Cancelled during an update while dependencies completed.
		return
	default:
	}

	// Run our callback or note that our upstream failed
	var response starriver.Response
	var upstreamFailed bool
	taskConfig := dataContext.Pipeline().GetTaskConfigure(v.ID())
	if depsSuccess {
		properties := make(map[string]interface{})
		dependNodes := make([]string, len(info.UpEdges))
		for idx, upEdge := range info.UpEdges {
			dependNodes[idx] = upEdge.Source().ID()
			for key, val := range upEdge.Properties() {
				properties[key] = val
			}
		}
		newDataContext, cancel := newNodeDataContext(dataContext, taskConfig.Timeout)
		if cancel != nil {
			defer cancel()
		}
		for _, upEdge := range info.UpEdges {
			newDataContext.AppendPrevTask(upEdge.Source().ID())
			newDataContext.AppendProperties(upEdge.Properties())
			if ce, ok := upEdge.(IsConditionalEdge); ok && !ce.Match(newDataContext) {
				// condition not match
				upstreamFailed = true
				response = helper.NewWarnResponse(fmt.Errorf("condition not match"))
				break
			}
		}
		if !upstreamFailed {
			if taskConfig.SkipExecution {
				dataContext.Pipeline().SetTaskStatus(v.ID(), starriver.TaskStatusSkipped)
				response = helper.NewSuccessResponse()
			} else {
				response = w.Callback(newDataContext, v)
				dataContext.Pipeline().SetTaskStatus(v.ID(), response.GetStatus())
				if data := response.GetData(); len(data) > 0 {
					newDataContext.SetCurrentNodeData(v.ID(), data)
				}

				if response.GetStatus() == starriver.TaskStatusBlocked {
					upstreamFailed = true
				} else if taskConfig.AlwaysPass {
					upstreamFailed = false
				}

				if response.GetFailureLevel() == starriver.FailureLevelFatal {
					dataContext.Stop()
				}
			}
		}
	} else {
		dataContext.Debugf("[TRACE] dag/walk: upstream of %q errored, so skipping", v.ID())
		upstreamFailed = true
		response = helper.NewWarnResponse(fmt.Errorf("upstream is failure"))
	}

	// Record the result (we must do this after execution because we mustn't
	// hold respLock while visiting a vertex.)
	w.respLock.Lock()
	if w.respMap == nil {
		w.respMap = make(map[string]starriver.Response)
	}
	w.respMap[v.ID()] = response
	if w.upstreamFailed == nil {
		w.upstreamFailed = make(map[string]struct{})
	}
	if upstreamFailed {
		w.upstreamFailed[v.ID()] = struct{}{}
	}
	w.respLock.Unlock()
}

func (w *Walker) waitDeps(
	dataContext starriver.DataContext,
	v Vertex,
	deps map[Vertex]<-chan struct{},
	doneCh chan<- bool,
	cancelCh <-chan struct{}) {

	// For each dependency given to us, wait for it to complete
	for dep, depCh := range deps {
	DepSatisfied:
		for {
			select {
			case <-depCh:
				// Dependency satisfied!
				dataContext.Debugf("Dependency satisfied. %q -> %q", dep.ID(), v.ID())
				break DepSatisfied

			case <-cancelCh:
				// Wait cancelled. Note that we didn't satisfy dependencies
				// so that anything waiting on us also doesn't run.
				doneCh <- false
				return
			}
		}
	}
	// Dependencies satisfied! We need to check if any errored
	w.respLock.Lock()
	defer w.respLock.Unlock()
	// 判断依赖的节点的状态
	var finalRes *struct{ pass bool }
	for dep := range deps {
		if resp := w.respMap[dep.ID()]; resp != nil {
			if n, ok := v.(Node); ok {
				switch n.GetType() {
				case NodeTypeAny:
					if resp.IsPass() {
						finalRes = &struct{ pass bool }{true}
						dataContext.Debugf("[TRACE] dag/walk: dependencies %q satisfied for AnyNode-> %q", dep.ID(), v.ID())
					} else {
						if finalRes == nil {
							finalRes = &struct{ pass bool }{false}
						}
						resp.SetFailureLevel(starriver.FailureLevelWarning)
						dataContext.Debugf("[TRACE] dag/walk: dependencies %q not satisfied for AnyNode, change to warning-> %q", dep.ID(), v.ID())
					}
				case NodeTypeNot:
					if resp.IsPass() {
						finalRes = &struct{ pass bool }{false}
					} else {
						finalRes = &struct{ pass bool }{true}
						resp.SetFailureLevel(starriver.FailureLevelWarning)
					}
				}
				continue
			}
			if !resp.IsPass() {
				dataContext.Debugf("[TRACE] dag/walk: dependencies %q not satisfied -> %q", dep.ID(), v.ID())
				finalRes = &struct{ pass bool }{false}
				break
			}
		}
	}
	if finalRes == nil {
		finalRes = &struct{ pass bool }{true}
	}
	doneCh <- finalRes.pass
	dataContext.Debugf("[TRACE] dag/walk: all dependencies result is %t -> %q", finalRes.pass, v.ID())
}

//func (w *Walker) node(nodeType NodeType, v Vertex,
//	deps map[Vertex]<-chan struct{},
//	doneCh chan<- bool) {
//	switch nodeType {
//	case NodeTypeAny:
//	case NodeTypeNot:
//	}
//}
