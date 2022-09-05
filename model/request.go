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
	ApiId      string     `json:"apiId" bson:"ApiId"`
	ApiName    string     `json:"apiName" bson:"ApiName"`
	URL        string     `json:"url" bson:"url"`
	Form       string     `json:"form" bson:"form"`     // http/webSocket/tcp/rpc
	Method     string     `json:"method" bson:"method"` // 方法 GET/POST/PUT
	Header     []*VarForm `json:"header" bson:"header"` // Headers
	Query      []*VarForm `json:"query" bson:"query"`
	Body       string     `json:"body" bson:"body"`
	Auth       []*VarForm `json:"auth" bson:"auth"`
	Parameters *sync.Map  `json:"parameters" bson:"parameters"`
	//Parameterizes      *sync.Map            `json:"parameterizes" bson:"parameterizes"`               // 接口中定义的变量
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

func ToString(varForm []*VarForm) (strs []string) {
	if varForm != nil {
		for _, v := range varForm {
			str := fmt.Sprintf("enable:%s name:%s value:%s valueType: %s", v.Enable, v.Name, v.Value, v.ValueType)
			strs = append(strs, str)
		}
	}
	return

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

func (r *Request) ReplaceUrlParameterizes(global *sync.Map) {

	urls := tools.FindAllDestStr(r.URL, "{{(.*?)}}")
	if urls != nil {
		for _, v := range urls {
			if value, ok := global.Load(v[1]); ok {
				r.URL = strings.Replace(r.URL, v[0], value.(string), -1)
			}
		}
	}
}

// ReplaceBodyParameterizes 替换body中的变量
func (r *Request) ReplaceBodyParameterizes() {
	bodyParameters := tools.FindAllDestStr(r.Body, "{{(.*?)}}")
	log.Logger.Error("body..................", bodyParameters)
	log.Logger.Error("global...............", r.Parameters)
	if bodyParameters != nil {
		for _, v := range bodyParameters {
			if value, ok := r.Parameters.Load(v[1]); ok {
				r.Body = strings.Replace(r.Body, v[0], value.(string), -1)
			}
		}
	}
}

// ReplaceQueryParameterizes 替换query中的变量
func (r *Request) ReplaceQueryParameterizes() {
	urls := tools.FindAllDestStr(r.URL, "{{(.*?)}}")
	if urls != nil {
		for _, v := range urls {
			if value, ok := r.Parameters.Load(v[1]); ok {
				r.URL = strings.Replace(r.URL, v[0], value.(string), -1)
			}
		}
	}
	bodys := tools.FindAllDestStr(r.Body, "{{(.*?)}}")
	if bodys != nil {
		for _, v := range bodys {
			if value, ok := r.Parameters.Load(v[1]); ok {
				r.Body = strings.Replace(r.Body, v[0], value.(string), -1)
			}
		}
	}
	if r.Query != nil {
		r.Query = ReplaceVarForm(r, r.Query)
	}
	if r.Header != nil {
		r.Header = ReplaceVarForm(r, r.Header)
	}
	if r.Auth != nil {
		r.Auth = ReplaceVarForm(r, r.Auth)
	}

}

func ReplaceVarForm(r *Request, varFormList []*VarForm) []*VarForm {
	if varFormList != nil {
		for _, queryVarForm := range varFormList {
			queryParameterizes := tools.FindAllDestStr(queryVarForm.Name, "{{(.*?)}}")
			if queryParameterizes != nil {
				for _, v := range queryParameterizes {
					if value, ok := r.Parameters.Load(v[1]); ok {
						queryVarForm.Name = strings.Replace(queryVarForm.Name, v[0], value.(string), -1)
					}
				}
			}
			queryParameterizes = tools.FindAllDestStr(queryVarForm.Value.(string), "{{(.*?)}}")
			if queryParameterizes != nil {
				for _, v := range queryParameterizes {
					if value, ok := r.Parameters.Load(v[1]); ok {
						queryVarForm.Value = strings.Replace(queryVarForm.Value.(string), v[0], value.(string), -1)
					}
				}
			}
		}
	}
	return varFormList
}

// FindParameterizes 将请求中的变量全部放到一个map中
func (r *Request) FindParameterizes() {
	if r.Parameters == nil {
		r.Parameters = new(sync.Map)
	}
	urls := tools.FindAllDestStr(r.URL, "{{(.*?)}}")
	for _, name := range urls {
		if _, ok := r.Parameters.Load(name[1]); !ok {
			r.Parameters.Store(name[1], name[0])
		}
	}
	bodyParameters := tools.FindAllDestStr(r.Body, "{{(.*?)}}")
	for _, name := range bodyParameters {
		if r.Parameters != nil {
			if _, ok := r.Parameters.Load(name[1]); !ok {
				r.Parameters.Store(name[1], name[0])
			}
		}

	}
	findVarFormParameters(r, r.Query)
	findVarFormParameters(r, r.Header)
	findVarFormParameters(r, r.Auth)
}

// ReplaceParameters 将场景变量中的值赋值给，接口变量
func (r *Request) ReplaceParameters(configuration *Configuration) {
	if r.Parameters == nil {
		r.Parameters = new(sync.Map)
	}
	r.Parameters.Range(func(k, v any) bool {
		if value, ok := configuration.Variable.Load(k); ok {
			if v == fmt.Sprintf("{{%s}}", k) {
				r.Parameters.Store(k, value)
			}
		}
		if _, ok := configuration.ParameterizedFile.VariableNames.VarMapList[k.(string)]; ok {
			if v == fmt.Sprintf("{{%s}}", k) {
				configuration.Mu.Lock()
				value := configuration.ParameterizedFile.UseVar(k.(string))
				r.Parameters.Store(k, value)
				configuration.Mu.Unlock()
			}
		}
		return true
	})
}

// 将请求中的[]VarForm中的变量，都存储到接口变量中
func findVarFormParameters(r *Request, varForms []*VarForm) {

	for _, varForm := range varForms {
		nameParameters := tools.FindAllDestStr(varForm.Name, "{{(.*?)}}")
		for _, name := range nameParameters {
			if _, ok := r.Parameters.Load(name[1]); !ok {
				r.Parameters.Store(name[1], name[0])
			}
		}
		valueParameters := tools.FindAllDestStr(varForm.Value.(string), "{{(.*?)}}")
		for _, value := range valueParameters {
			if len(value) > 1 {
				if _, ok := r.Parameters.Load(value[1]); !ok {
					r.Parameters.Store(value[1], value[0])
				}
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
