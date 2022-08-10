package request

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"kp-runner/model"
	"strconv"
	"sync"
	"time"
)

// Form 支持协议
const (
	FormTypeHTTP      = "http"
	FormTypeHTTPS     = "https"
	FormTypeWebSocket = "webSocket"
	FormTypeGRPC      = "grpc"
)

// 断言类型
const (
	Text = iota
	Regular
	Json
	XPath
)

// 文本断言类型
const (
	ResponseCode = iota
	ResponseHeaders
	ResponseData

	Contain         = "包含"
	NotContain      = "不包含"
	Equal           = "等于"
	NotEqual        = "不等于"
	OriginatingFrom = "以...开始"
	EndIn           = "以...结束"
)

// Assertion 断言
type Assertion struct {
	Type             int8             `json:"type"` //  0:Text; 1:Regular; 2:Json; 3:XPath
	AssertionText    AssertionText    `json:"assertionText"`
	AssertionRegular AssertionRegular `json:"assertionRegular"`
	AssertionJson    AssertionJson    `json:"assertionJson"`
	AssertionXPath   AssertionXPath   `json:"assertionXPath"`
}

// AssertionText 文本断言 0
type AssertionText struct {
	AssertionTarget int8        `json:"AssertionTarget"` // 0: ResponseCode; 1:ResponseHeaders; 2:ResponseData
	Condition       string      `json:"condition"`       // Contain、NotContain、Equal、NotEqual、OriginatingFrom、EndIn
	Value           interface{} `json:"value"`
}

// AssertionRegular 正则断言 1
type AssertionRegular struct {
	AssertionTarget int8   `json:"type"`       // 2:ResponseData
	Expression      string `json:"expression"` // 正则表达式

}

// AssertionJson json断言 2
type AssertionJson struct {
	Expression string `json:"expression"` // json表达式
	Condition  string `json:"condition"`  // Contain、NotContain、Equal、NotEqual
}

// AssertionXPath xpath断言 3
type AssertionXPath struct {
}

// Request 请求数据
type Request struct {
	ApiId                int               `json:"apiId"`
	ApiName              string            `json:"apiName"`
	URL                  string            `json:"url"`
	Form                 string            `json:"form"`    // http/webSocket/tcp/rpc
	Method               string            `json:"method"`  // 方法 GET/POST/PUT
	Headers              map[string]string `json:"headers"` // Headers
	Body                 string            `json:"body"`
	Parameterizes        map[string]string `json:"parameterizes"`        // 接口中定义的变量
	Assertions           []Assertion       `json:"assertions"`           // 验证的方法(断言)
	Timeout              time.Duration     `json:"timeout"`              // 请求超时时间
	ErrorThreshold       float64           `json:"errorThreshold"`       // 错误率阈值
	CustomRequestTime    uint8             `json:"customRequestTime"`    // 自定义响应时间线
	RequestTimeThreshold uint8             `json:"requestTimeThreshold"` // 响应时间阈值
	Debug                bool              `json:"debug"`                // 是否开启Debug模式
	Connection           int               `json:"connection"`           // 0:websocket长连接
	Weight               int8              `json:"weight"`               // 权重，并发分配的比例
	Tag                  bool              `json:"tag"`                  // Tps模式下，该标签代表以该接口为准
}

// VerifyAssertionText 验证断言
func (assertionText AssertionText) VerifyAssertionText(response *fasthttp.Response) (code int, msg string) {
	switch assertionText.AssertionTarget {
	case ResponseCode:
		switch assertionText.Condition {
		case Equal:
			if assertionText.Value == response.StatusCode() {
				return model.NoError, strconv.Itoa(response.StatusCode()) + Equal + strconv.Itoa(assertionText.Value.(int)) + "成功"
			} else {
				return model.AssertError, strconv.Itoa(response.StatusCode()) + Equal + strconv.Itoa(assertionText.Value.(int)) + "失败"
			}
		case NotEqual:
			if assertionText.Value == response.StatusCode() {
				return model.NoError, strconv.Itoa(response.StatusCode()) + Equal + strconv.Itoa(assertionText.Value.(int)) + "成功"
			} else {
				return model.AssertError, strconv.Itoa(response.StatusCode()) + Equal + strconv.Itoa(assertionText.Value.(int)) + "失败"

			}
		}
	case ResponseHeaders:
		switch assertionText.Condition {
		case Contain:

		}
	case ResponseData:

	}
	return
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
