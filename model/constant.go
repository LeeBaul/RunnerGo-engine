package model

// Form 支持协议类型
const (
	FormTypeHTTP      = "api"       // http协议
	FormTypeWebSocket = "websocket" // webSocket协议
	FormTypeGRPC      = "grpc"      // grpc协议
)

// 返回 code 码
const (
	// NoError 没有错误
	NoError = int64(10000)
	// AssertError 断言错误
	AssertError = int64(10001)
	// RequestError 请求错误
	RequestError = int64(10002)
	// ServiceError 服务错误
	ServiceError = int64(10003)
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
	RequestType    = "request"    // 接口请求
	ControllerType = "controller" // 控制器

	IfControllerType   = "if"         // if控制器
	WaitControllerType = "wait"       // 等待控制器
	CollectionType     = "collection" // 集合点控制器

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

// 数据类型
const (
	StringType    = "String"
	TextType      = "Text"
	ObjectType    = "Object"
	ArrayType     = "Array"
	IntegerType   = "Integer"
	NumberType    = "Number"
	FloatType     = "Float"
	DoubleType    = "Double"
	FileType      = "File"
	DateType      = "Date"
	DateTimeType  = "DateTime"
	TimeStampType = "TimeStampType"
	BooleanType   = "boolean"
)

const (
	KVType     = "kv"
	BearerType = "bearer"
	BasicType  = "basic"
)

const (
	NoneMode      = "none"
	FormMode      = "form-data"
	UrlencodeMode = "x-www-form-urlencoded"
	JsonMode      = "json"
	XmlMode       = "xml"
	JSMode        = "javascript"
	PlainMode     = "plain"
	HtmlMode      = "html"
)
