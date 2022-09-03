package model

import (
	"fmt"
	"kp-runner/log"
	"kp-runner/model/proto/pb"
	"kp-runner/tools"
	"strings"
	"sync"
)

// Request 请求数据
type Request struct {
	ApiId   string    `json:"apiId" bson:"ApiId"`
	ApiName string    `json:"apiName" bson:"ApiName"`
	URL     string    `json:"url" bson:"url"`
	Form    string    `json:"form" bson:"form"`     // http/webSocket/tcp/rpc
	Method  string    `json:"method" bson:"method"` // 方法 GET/POST/PUT
	Header  []VarForm `json:"header" bson:"header"` // Headers
	Query   []VarForm `json:"query" bson:"query"`
	Body    string    `json:"body" bson:"body"`
	Auth    []VarForm `json:"auth" bson:"auth"`
	//Parameterizes        *sync.Map            `json:"parameterizes" bson:"parameterizes"`               // 接口中定义的变量
	Assertions           []*Assertion         `json:"assertions" bson:"assertions"`                     // 验证的方法(断言)
	Timeout              int64                `json:"timeout" bson:"timeout"`                           // 请求超时时间
	ErrorThreshold       float32              `json:"errorThreshold" bson:"errorThreshold"`             // 错误率阈值
	CustomRequestTime    int64                `json:"customRequestTime" bson:"customRequestTime"`       // 自定义响应时间线
	RequestTimeThreshold int64                `json:"requestTimeThreshold" bson:"requestTimeThreshold"` // 响应时间阈值
	Regulars             []*RegularExpression `json:"regulars" bson:"regulars"`                         // 正则表达式
	Debug                bool                 `json:"debug" bson:"debug"`                               // 是否开启Debug模式
	Connection           int64                `json:"connection" bson:"connection"`                     // 0:websocket长连接
	Weight               int64                `json:"weight" bson:"weight"`                             // 权重，并发分配的比例
	Tag                  bool                 `json:"tag" bson:"tag"`                                   // Tps模式下，该标签代表以该接口为准
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

// VarForm 参数表
type VarForm struct {
	Enable    bool        `json:"enable" bson:"enable"`
	Name      string      `json:"name" bson:"name"`
	Value     interface{} `json:"value" bson:"value"`
	ValueType string      `json:"valueType" bson:"valueType"`
}

// Conversion 将其他类型转换为string
func (v VarForm) Conversion() {
	switch v.ValueType {
	case StringType:
		// 字符串类型不用转换
	case TextType:
		// 文本类型不用转换
	case ObjectType:
		// 对象不用转换
	case ArrayType:
		// 数组不用转换
	case IntegerType:
		v.Value = fmt.Sprintf("%d", v.Value)
	case NumberType:
		v.Value = fmt.Sprintf("%d", v.Value)
	case FloatType:
		v.Value = fmt.Sprintf("%f", v.Value)
	case DoubleType:
		v.Value = fmt.Sprintf("%f", v.Value)
	case FileType:
	case DateType:
	case DateTimeType:
	case TimeStampType:
	case BooleanType:

	}
}

func (r Request) ReplaceUrlParameterizes(global *sync.Map) {

	urls := tools.FindAllDestStr(r.URL, "{{(.*?)}}")
	if urls != nil {
		for _, v := range urls {
			if value, ok := global.Load(v[1]); ok {
				r.URL = strings.Replace(r.URL, v[0], value.(string), -1)
			}
		}
	}
}

func (r Request) VariableSubstitution(global *sync.Map) {
	for _, varForm := range r.Header {
		// 查找header的key中是否存在变量{{****}}
		keys := tools.FindAllDestStr(varForm.Name, "{{(.*?)}}")
		if keys != nil {
			for _, realName := range keys {
				if value, ok := global.Load(realName[1]); ok {
					varForm.Name = strings.Replace(varForm.Name, realName[0], value.(string), -1)
				}
			}
		}

		values := tools.FindAllDestStr(varForm.Value.(string), "{{(.*?)}}")
		if values != nil {
			for _, realValue := range values {
				if value, ok := global.Load(realValue[1]); ok {
					varForm.Value = strings.Replace(varForm.Value.(string), realValue[0], value.(string), -1)
				}
			}
		}
	}
}

func (r *Request) ReplaceBodyParameterizes(global *sync.Map) {
	bodys := tools.FindAllDestStr(r.Body, "{{(.*?)}}")
	log.Logger.Error("body..................", bodys)
	log.Logger.Error("global...............", global)
	if bodys != nil {
		for _, v := range bodys {
			if value, ok := global.Load(v[1]); ok {
				r.Body = strings.Replace(r.Body, v[0], value.(string), -1)
			}
		}
	}
}

func GrpcReplaceParameterizes(r *pb.Request, globalVariable *sync.Map) {
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

		values := tools.FindAllDestStr(v.String(), "{{(.*?)}}")
		if values != nil {
			for _, realValue := range values {
				if value, ok := globalVariable.Load(realValue[1]); ok {
					v.Value = []byte(strings.Replace(v.String(), realValue[0], value.(string), -1))
				}
			}
			r.Parameterizes[k] = v
		}
	}
}
