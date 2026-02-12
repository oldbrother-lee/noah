package initializer

// 初始化顺序常量
// 数字越小越先执行
const (
	// 基础表结构
	InitOrderRole     = 100
	InitOrderUser     = 200
	InitOrderMenu     = 300
	InitOrderAPI      = 400
	InitOrderDept     = 500
	
	// 关联关系
	InitOrderUserRole = 600
	InitOrderRBAC     = 700
	
	// 业务功能
	InitOrderFlow     = 800
	InitOrderInspect  = 900
	
	// Insight 功能
	InitOrderInsight = 1000
)
