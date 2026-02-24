package source

// ApiSeedItem 与 api 表字段对应，用于种子数据与初始化插入
type ApiSeedItem struct {
	Group  string `json:"group"`
	Name   string `json:"name"`
	Path   string `json:"path"`
	Method string `json:"method"`
}
