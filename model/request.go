package model

import (
	"kp-runner/tools"
	"strings"
	"sync"
)

// Request 请求数据
type Request struct {
	ApiId                string               `json:"apiId"`
	ApiName              string               `json:"apiName"`
	URL                  string               `json:"url"`
	Form                 string               `json:"form"`    // http/webSocket/tcp/rpc
	Method               string               `json:"method"`  // 方法 GET/POST/PUT
	Headers              map[string]string    `json:"headers"` // Headers
	Body                 string               `json:"body"`
	Requests             []*Request           `json:"requests"`
	Controllers          []*Controller        `json:"controllers"`
	Parameterizes        map[string]string    `json:"parameterizes"`        // 接口中定义的变量
	Assertions           []*Assertion         `json:"assertions"`           // 验证的方法(断言)
	Timeout              int                  `json:"timeout"`              // 请求超时时间
	ErrorThreshold       float64              `json:"errorThreshold"`       // 错误率阈值
	CustomRequestTime    uint8                `json:"customRequestTime"`    // 自定义响应时间线
	RequestTimeThreshold uint8                `json:"requestTimeThreshold"` // 响应时间阈值
	Regulars             []*RegularExpression `json:"regulars"`             // 正则表达式
	Debug                bool                 `json:"debug"`                // 是否开启Debug模式
	Connection           int                  `json:"connection"`           // 0:websocket长连接
	Weight               int8                 `json:"weight"`               // 权重，并发分配的比例
	Tag                  bool                 `json:"tag"`                  // Tps模式下，该标签代表以该接口为准
}

type RegularExpression struct {
	VariableName string `json:"variableName"` // 变量
	Expression   string `json:"expression"`   // 表达式
}

// Extract 提取response 中的值
func (re RegularExpression) Extract(str string, parameters *sync.Map) {
	name := tools.VariablesMatch(re.VariableName)
	if value := tools.FindDestStr(str, re.Expression); value != "" {
		parameters.Store(name, value)
	}
}

func (r *Request) ReplaceUrlParameterizes() {

	urls := tools.FindAllDestStr(r.URL, "{{(.*?)}}")
	if urls != nil {
		for _, v := range urls {
			if value, ok := r.Parameterizes[v[1]]; ok {
				r.URL = strings.Replace(r.URL, v[0], value, -1)
			}
		}
	}
}

func (r *Request) ReplaceHeaderParameterizes() {
	for k, v := range r.Headers {

		// 查找header的key中是否存在变量{{****}}
		keys := tools.FindAllDestStr(k, "{{(.*?)}}")
		if keys != nil {
			delete(r.Headers, k)
			for _, realKey := range keys {
				if value, ok := r.Parameterizes[realKey[1]]; ok {
					k = strings.Replace(k, realKey[0], value, -1)
				}
			}
			r.Headers[k] = v
		}

		values := tools.FindAllDestStr(v, "{{(.*?)}}")
		if values != nil {
			for _, realValue := range values {
				if value, ok := r.Parameterizes[realValue[1]]; ok {
					v = strings.Replace(v, realValue[0], value, -1)
				}
			}
			r.Headers[k] = v
		}
	}
}

func (r *Request) ReplaceParameterizes(globalVariable *sync.Map) {
	for k, v := range r.Parameterizes {
		// 查找header的key中是否存在变量{{****}}
		keys := tools.FindAllDestStr(k, "{{(.*?)}}")
		if keys != nil {
			delete(r.Parameterizes, k)
			for _, realKey := range keys {
				if value, ok := globalVariable.Load(realKey[1]); ok {
					k = strings.Replace(k, realKey[0], value.(string), -1)
				}
			}
			r.Parameterizes[k] = v
		}

		values := tools.FindAllDestStr(v, "{{(.*?)}}")
		if values != nil {
			for _, realValue := range values {
				if value, ok := globalVariable.Load(realValue[1]); ok {
					v = strings.Replace(v, realValue[0], value.(string), -1)
				}
			}
			r.Parameterizes[k] = v
		}
	}
}

func (r *Request) ReplaceBodyParameterizes() {
	bodys := tools.FindAllDestStr(r.Body, "{{(.*?)}}")
	if bodys != nil {
		for _, v := range bodys {
			if value, ok := r.Parameterizes[v[1]]; ok {
				r.Body = strings.Replace(r.Body, v[0], value, -1)
			}
		}
	}
}
