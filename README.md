# starriver
StarRiver（
星河）是一个基于 DAG 的流程编排 SDK。 用户可以利用系统预置的节点以及自定义的节点组织流程。通过工作台、节点输出以及边的属性等获取组装参数，最终完成整个流程。

## 基本概念
* 流程（Pipeline）：即整个流程，由一个又一个的任务根据依赖关系组合而成。
* 任务（Task）：每个流程中的一个节点，执行某个业务逻辑。可能有直接输出也可能没有输出。结果可以放置到工作台上，供其他任意任务使用。
* 节点（Node）：流程中的节点，系统当前提供了两种特殊的内置节点。Any（任意依赖满足即满足）节点，Fail（当且仅当依赖节点 not pass 才满足）节点。
* 组件（Component）：用户自定义的任务节点，通过实现一系列特殊的接口即可自由添加到流程中。
* 边（Edge）：流程中的依赖关系。每条边都有方向，Source → Target，表示 Target依赖 Source 成功执行才可执行。同时每条边还可以携带属性。
* 工作台（DataContext）：流程中的节点想要获取数据时，从工作台获取；若有输出供其他节点使用，也需要将其设置到工作台中。

## 设计思想

![1.png](docs%2F1.png)

## 逻辑关系
### ANY 或关系
当某个（或多个）task 有多个前置的依赖 task，且它并不要求所有的前置依赖都成功，只要有一个成功即可执行。那可以使用系统提供的 Any 节点，这个节点不做任何的业务逻辑，只在依赖的节点有任一成功时，将自身也标记为成功，以便流程继续执行。
![2.png](docs%2F2.png)

### AND 与关系
多个依赖必须全部满足时，系统默认支持，未引入任何内置节点。
![3.png](docs%2F3.png)

### Not 非关系
当依赖的节点返回结果是 not pass 时才会执行，可用于构建 if/else 或者降级等场景。
![4.png](docs%2F4.png)
如上述逻辑为:
```
// (A || C)&&D
if A then
   D
else if C then
   D
```
## 执行数据获取
每个节点获取数据的地方有三个。
* 边属性
* 依赖节点输出
* 工作台共享数据
![5.png](docs%2F5.png)

每个节点可以通过 DataContext 获取指定的数据，整个数据获取流程如下：
1. 数据会优先从依赖的边的属性获取，若未获取，则进入2
2. 从上个节点的输出结果中获取，若未获取，则进入3
3. 从工作台共享数据中获取。

TIPS：当需要 mock 数据测试流程时，可以通过边属性实现。

## 条件选择
有些场景，需要有分支选择能力。如根据上一个节点的输出，决定是否进入下一个节点执行。这时候，我们可以在边上增加条件属性。
```
condition: # 条件，当且仅当 key 获得的值等于这里的 value 时，才会进入此节点执行
   key: user_level
   value: [C4, C5]
   type: in 
```
如上述示例中，当 user_level 是 C4 或 C5 时，才会进行下一个节点。否则将认为依赖不满足。user_level 的取值逻辑遵从上述逻辑（边->输出->共享数据）。条件关系支持以下逻辑操作符：
**in, == , >, <, >=, <=, !=**

## 自定义组件
若需要定义自己的组件，只需要实现下面接口即可。
```go
type Executable interface {
	ID() string
	Execute(DataContext, interface{}) Response
}
```
其中，最重要的方法是 Execute 方法，它接收了两个参数，第一个参数是 DataContext，我们可以通过它去获取想要的数据，记录日志等等。第二个参数是自定义的参数（若需要的话，否则将是 nil），系统会根据配置给我们获取数据组装好参数。在 Execute 方法里面直接转换给给定的参数即可。

要使用自定义参数，我们需要实现 WithParameters 接口，让它返回我们的自定义参数结构体。
```go
type WithParameters interface {
    Parameters() interface{}
}
```
下面我们给一个示例。
```go
type ExprParam struct {
	Expr string `desc:"规则表达式" required:"true"`
}


func (e *expression) ParameterNew() interface{} {
	return &ExprParam{}
}


func (e *expression) Execute(dataContext types.DataContext, param interface{}) types.Response {
	exprParam := params.(*ExprParam)
	...
}
```
参数配置支持三种类型：literal（字面量），variable（变量），complex（复合类型）, mapping（映射）
其中
* 字面量是固定值，该节点每次运行都使用该值，需要在配置流程时明确给出；
* 变量需要给出变量名，将通过它去工作台中获取运行时的值（可能是上一个节点的输出或者其他节点写入的值）；
* 复合类型是前两种类型参数的组合，它可以递归定义，最终会输出一个参数数组，数组中每个参数可以是字面量或变量任意组合而成。
* 映射类型对应 Golang 的 map，给定 map 的key，它的值可以是字面量、变量或复合类型（甚至是映射类型）的任何一种。由于值不固定，需要将参数类型设置为 map[string]interface{}

另外，我们如果想要在执行前后运行特定的业务逻辑，我们可以让自定义组件实现下面两个接口
```go
BeforeExecute interface {
	Before(dataContext DataContext)
}

AfterExecute interface {
	After(dataContext DataContext)
}
```
最后，我们可以在运行后根据执行结果获得回调，那可以实现 Listener 接口
```go
Listener interface {
	OnSuccess(dataContext DataContext, result map[string]interface{})
	OnFailure(dataContext DataContext, err error)
}
```

## DEMO
1、我们可以通过 yaml （或者json）定义一个流程
```yaml
# 流程名称
name: workflow_name
# 并发数，默认 10
concurrency: 10
# 超时时间
timeout: 1m
# 环境变量，组件里可以通过 DataContext.Env 获得值，执行过程中不可更改
env:
  app: demo_app
  author: jimmy
# 如果流程最终需要返回值，这里可以指定数据的 key，最终返回 map
result:
  - result_key_a
  - result_key_b
pipeline:
  - task: task123  #task id，同一个流程里面一个 Component 可以重复多次，但是 id 是唯一的。
    namespace: antispam # Component 的命名空间，只会在该命名空间下查找 Component，如果不需要则留空
    name: CustomComponent1 # Component 名
    # task 配置项
    config:
      timeout: 1s # 执行超时时间
      always_pass: true # 无论执行结果如何，最终节点都成功
      skip_execution: true # 跳过实际执行，直接返回成功
      abort_if_error: true # 当节点执行有错误时，中断整个流程的执行
      params:
        -
          name: XXX # 对应组件参数 struct 的名称
          type: variable # 值数据类型，根据 variable 从 dataContext 中获取数据
          variable: abc # key 的名称
          required: false
        -
          name: xxx
          type: literal  # 字面量，从 literal 里取值
          literal: testValue # 参数实际的值
          required: true
        -
          name: xxx2
          type: complex # 复合类型，此类型最终会组成一个参数数组，complex 里是每一个元素的具体的取值
          required: true
          complex:
            -
              type: variable
              variable: key1
              required: true
            -
              type: literal
              literal: 123
        -
          name: xxx3
          type: mapping # 映射类型
          required: false
          mapping:
              key1:
                type: literal
                literal: 1228
              key2:
                type: veriable
                veriable: xxx
  - task: task_id_456
    name: Expression
    config:
      params:
        -
          name: Expr
          type: variable
          value: age > 18
          required: true
    depends: #依赖的节点
      - task: task123 # 依赖的前置节点 id
        condition: # 条件，当且仅当 key 获得的值等于这里的 value 时，才会进入此节点执行
          key: user_level
          value: [C4, C5]
          type: in # 可选比较符有：「in, == , >, <, >=, <=, !=」
        properties: # 属性，会带入到此节点中，优先级最高
          key1: value1
          key2: value2
```
2、执行它
```go
// 加载 yaml 配置文件，获得流程配置
conf, err := flow.LoadPipelineByYaml(flowStr)
// 创建一个流程实例
pipeline, err := flow.NewPipeline(conf)
// 创建数据上下文
dataContext := flow.NewContext(context.TODO(), pipeline)
// 创建一个流程引擎实例
re := flow.NewRiverEngine()
defer re.Destroy()

 // 执行流程，获得结果
result, err := re.Run(dataContext, pipeline)
/* result 结构体示例
	Result struct {
		Data     map[string]interface{}  // 若成功且流程中指明需要某些key做为结果则将设置在此
		Snapshot []byte                  // 若流程执行中断 blocked，则将保存至此，后续如果需要恢复执行，则必须自行保存
		Status   PipelineStatus          // 流程的最终执行状态
		State    map[string]TaskStatus   // 各个节点的执行结果，如果是 blocked，则需要自行保存，后续恢复执行需要
		Error    error                   // 执行过程中发生的错误
	}
*/

if result.Status == starriver.PipelineStatusBlocked {
    // 快照将按照预设的序列化方式进行，需要自行保存。
	dataSnapshot := result.Snapshot
	state := result.State
	// 等待 blocked 资源恢复后，blocked 和 init 的 task 将会重跑
    dataStore := unmarshal(dataSnapshot)
    dataContext, pipeline := flow.Rebuild(ctx, pipelineConf, state, dataStore, initialData)
        
    flow.NewRiverEngine().Run(dataContext, pipeline)
}
```
自定义组件示例
```go
import (
	"reflect" 

    "github.com/thanksloving/starriver/registry"
	"github.com/thanksloving/starriver/types"
)

type (
	customComponent struct {
		id string
	}
	customParam struct {
		Name   string 
		Age    int    
		Gender int    `default:"1"`  // 参数可以有默认值
	}
) 

func (c *customComponent) ID() string {
	return c.id
}

func (c *customComponent) Execute(dataContext types.DataContext, param interface{}) types.Response {
	person := param.(*customParam)
	dataContext.Infof("welcome %+v", person) // 打印日志
	dataContext.Set(dataContext.Context(), "Name", person.Name)  // 保存数据到工作台，其他组件可以使用
	return types.NewPassResponse(map[string]interface{}{"result": person.Age}) // 输出结果，下一个节点可以获取   
}

func (c *customComponent) ParameterNew() interface{} {
	return &customParam{}
}
```
注册
```go
import (
	"github.com/thanksloving/starriver/registry"
	"github.com/thanksloving/starriver/types"
)


// 将创建组件的方法在合适的地方（使用前即可）注册到组件库中。 
func init() {
    registry.Register("CustomComponentXXX", "这里是描述",
		func(id string) types.Executable {
			return &customComponent{
				id: id,
			}
		},
		registry.Input([]types.InputParam{
			{
				Key:      "Name",
				Required: true,
				Desc:     "姓名",
			},
            {
				Key:      "Age",
				Required: true,
				Desc:     "年龄",
			},
            {
				Key:      "Gender",
				Required: true,
				Desc:     "性别",
			},
		}),
		registry.Output(map[string]types.OutputValue{
			"result": {
				Desc: "年龄",
                Type: reflect.Int,
			},
		}),
        registry.Namespace("test"),
	)
}
```

