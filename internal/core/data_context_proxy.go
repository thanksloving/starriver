package core

func (dc *dataContext) Set(key string, val interface{}) bool {
	return dc.SharedDataStore.Set(dc, key, val)
}

func (dc *dataContext) Del(key string) {
	dc.SharedDataStore.Del(dc, key)
}

func (dc *dataContext) Get(key string) (interface{}, bool) {
	return dc.SharedDataStore.Get(dc, key)
}

func (dc *dataContext) Marshal() ([]byte, error) {
	return dc.SharedDataStore.Marshal()
}

func (dc *dataContext) Configure(flowName, requestID string) {
	dc.SharedDataStore.Configure(flowName, requestID)
}

func (dc *dataContext) SetCurrentNodeData(nodeId string, data map[string]interface{}) {
	dc.SharedDataStore.SetCurrentNodeData(dc, nodeId, data)
}

func (dc *dataContext) GetDependNodeValue(nodeId, key string) (interface{}, bool) {
	return dc.SharedDataStore.GetDependNodeValue(dc, nodeId, key)
}

func (dc *dataContext) Debug(msg string) {
	dc.Logger.Debug(dc, msg)
}

func (dc *dataContext) Debugf(msg string, args ...interface{}) {
	dc.Logger.Debugf(dc, msg, args...)
}

func (dc *dataContext) Info(msg string) {
	dc.Logger.Info(dc, msg)
}

func (dc *dataContext) Infof(msg string, args ...interface{}) {
	dc.Logger.Infof(dc, msg, args...)
}

func (dc *dataContext) Warn(msg string) {
	dc.Logger.Warn(dc, msg)
}

func (dc *dataContext) Warnf(msg string, args ...interface{}) {
	dc.Logger.Warnf(dc, msg, args...)
}

func (dc *dataContext) Error(msg string) {
	dc.Logger.Error(dc, msg)
}

func (dc *dataContext) Errorf(msg string, args ...interface{}) {
	dc.Logger.Errorf(dc, msg, args...)
}

func (dc *dataContext) Fatal(msg string) {
	dc.Logger.Fatal(dc, msg)
}

func (dc *dataContext) Fatalf(msg string, args ...interface{}) {
	dc.Logger.Fatalf(dc, msg, args...)
}
