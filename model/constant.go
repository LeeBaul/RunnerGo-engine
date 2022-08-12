package model

// Form 支持协议类型
const (
	FormTypeHTTP      = "http"      // http协议
	FormTypeHTTPS     = "https"     // https协议
	FormTypeWebSocket = "webSocket" // webSocket协议
	FormTypeGRPC      = "grpc"      // grpc协议
)

// 返回 code 码
const (
	// NoError 没有错误
	NoError = 10000
	// AssertError 断言错误
	AssertError = 10001
	// RequestError 请求错误
	RequestError = 10002
	// ServiceError 服务错误
	ServiceError = 10003
)

// 断言类型
const (
	Text    = iota // 文本断言
	Regular        // 正则表达式
	Json           // json断言
	XPath          // xpath断言
)

// 文本断言类型
const (
	ResponseCode    = iota // 断言响应码
	ResponseHeaders        // 断言响应的信息头
	ResponseData           // 断言响应的body信息
)

// 事件类型
const (
	IfControllerType   = "if"         // if控制器
	WaitControllerType = "wait"       // 等待控制器
	CollectionType     = "collection" // 集合点控制器
	RequestType        = "api"        // 接口请求
)

// 逻辑运算符
const (
	Equal              = "eq"         // 等于
	UNEqual            = "uneq"       // 不等于
	GreaterThan        = "gt"         // 大于
	GreaterThanOrEqual = "qte"        // 大于或等于
	LessThan           = "lt"         // 小于
	LessThanOrEqual    = "lte"        // 小于或等于
	Includes           = "includes"   // 包含
	UNIncludes         = "unincludes" // 不包含
	NULL               = "null"       // 为空
	NotNULL            = "notnull"    // 不为空

	OriginatingFrom = "以...开始"
	EndIn           = "以...结束"
)
