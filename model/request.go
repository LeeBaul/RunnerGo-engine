package model

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"kp-runner/tools"
	"sync"
)

// Request 请求数据
type Request struct {
	ApiId                string              `json:"apiId"`
	ApiName              string              `json:"apiName"`
	URL                  string              `json:"url"`
	Form                 string              `json:"form"`    // http/webSocket/tcp/rpc
	Method               string              `json:"method"`  // 方法 GET/POST/PUT
	Headers              map[string]string   `json:"headers"` // Headers
	Body                 string              `json:"body"`
	Parameterizes        map[string]string   `json:"parameterizes"`        // 接口中定义的变量
	Assertions           []Assertion         `json:"assertions"`           // 验证的方法(断言)
	Timeout              int                 `json:"timeout"`              // 请求超时时间
	ErrorThreshold       float64             `json:"errorThreshold"`       // 错误率阈值
	CustomRequestTime    uint8               `json:"customRequestTime"`    // 自定义响应时间线
	RequestTimeThreshold uint8               `json:"requestTimeThreshold"` // 响应时间阈值
	Regulars             []RegularExpression `json:"regulars"`
	Debug                bool                `json:"debug"`      // 是否开启Debug模式
	Connection           int                 `json:"connection"` // 0:websocket长连接
	Weight               int8                `json:"weight"`     // 权重，并发分配的比例
	Tag                  bool                `json:"tag"`        // Tps模式下，该标签代表以该接口为准
}

type RegularExpression struct {
	VariableName string `json:"variableName"` // 变量
	Expression   string `json:"expression"`   // 表达式
}

// Extract 提取response 中的值
func (re RegularExpression) Extract(str string, parameters map[string]string) {
	name := tools.VariablesMatch(re.VariableName)
	if value := tools.FindDestStr(str, re.Expression); value != "" {
		parameters[name] = value
	}
}

// 校验函数
var (
	// verifyMapHTTP http 校验函数
	verifyMapHTTP = make(map[string]VerifyHTTP)
	// verifyMapHTTPMutex http 并发锁
	verifyMapHTTPMutex sync.RWMutex
	// verifyMapWebSocket webSocket 校验函数
	verifyMapWebSocket = make(map[string]VerifyWebSocket)
	// verifyMapWebSocketMutex webSocket 并发锁
	verifyMapWebSocketMutex sync.RWMutex
)

// RegisterVerifyHTTP 注册 http 校验函数
func RegisterVerifyHTTP(verify string, verifyFunc VerifyHTTP) {
	verifyMapHTTPMutex.Lock()
	defer verifyMapHTTPMutex.Unlock()
	key := fmt.Sprintf("%s.%s", FormTypeHTTP, verify)
	verifyMapHTTP[key] = verifyFunc
}

// RegisterVerifyWebSocket 注册 webSocket 校验函数
func RegisterVerifyWebSocket(verify string, verifyFunc VerifyWebSocket) {
	verifyMapWebSocketMutex.Lock()
	defer verifyMapWebSocketMutex.Unlock()
	key := fmt.Sprintf("%s.%s", FormTypeWebSocket, verify)
	verifyMapWebSocket[key] = verifyFunc
}

// VerifyHTTP http 验证
type VerifyHTTP func(request *Request, response *fasthttp.Response) (code int, isSucceed bool)

// VerifyWebSocket webSocket 验证
type VerifyWebSocket func(request *Request, seq string, msg []byte) (code int, isSucceed bool)

//// getVerifyKey 获取校验 key
//func (r *Request) getVerifyKey() (key string) {
//	return fmt.Sprintf("%s.%s", r.Form, r.Verify)
//}
//
//// GetVerifyHTTP 获取数据校验方法
//func (r *Request) GetVerifyHTTP() VerifyHTTP {
//	verify, ok := verifyMapHTTP[r.getVerifyKey()]
//	if !ok {
//		panic("GetVerifyHTTP 验证方法不存在:" + r.Verify)
//	}
//	return verify
//}
//
//// GetVerifyWebSocket 获取数据校验方法
//func (r *Request) GetVerifyWebSocket() VerifyWebSocket {
//	verify, ok := verifyMapWebSocket[r.getVerifyKey()]
//	if !ok {
//		panic("GetVerifyWebSocket 验证方法不存在:" + r.Verify)
//	}
//	return verify
//}

func (r *Request) GetBody() (body []byte) {
	return []byte(r.Body)
}
