package model

import (
	"encoding/json"
	"fmt"
	"kp-runner/model/proto/pb"
	"kp-runner/tools"
	"strconv"
	"strings"
	"sync"
)

// Api 请求数据
type Api struct {
	TargetId   int64                `json:"target_id" bson:"target_id"`
	Name       string               `json:"name" bson:"name"`
	TargetType string               `json:"target_type" bson:"target_type"` // api/webSocket/tcp/grpc
	Method     string               `json:"method" bson:"method"`           // 方法 GET/POST/PUT
	Request    Request              `json:"request" bson:"request"`
	Parameters *sync.Map            `json:"parameters" bson:"parameters"`
	Assert     []*AssertionText     `json:"assert" bson:"assert"`         // 验证的方法(断言)
	Timeout    int64                `json:"timeout" bson:"timeout"`       // 请求超时时间
	Regex      []*RegularExpression `json:"regex" bson:"regex"`           // 正则表达式
	Debug      bool                 `json:"debug" bson:"debug"`           // 是否开启Debug模式
	Connection int64                `json:"connection" bson:"connection"` // 0:websocket长连接

}

type Request struct {
	URL    string  `json:"url" bson:"url"`
	Header *Header `json:"header" bson:"header"` // Headers
	Query  *Query  `json:"query" bson:"query"`
	Body   *Body   `json:"body" bson:"body"`
	Auth   *Auth   `json:"auth" bson:"auth"`
	Cookie *Cookie `json:"cookie" bson:"cookie"`
}

type Body struct {
	Mode      string     `json:"mode" bson:"mode"`
	Raw       string     `json:"raw" bson:"raw"`
	Parameter []*VarForm `json:"parameter" bson:"parameter"`
}

func (b *Body) ToString() (s string) {
	if b == nil {
		return s
	}
	switch b.Mode {
	case NoneMode:
	case FormMode:
		body := make(map[string]interface{})
		for _, value := range b.Parameter {
			if value.IsChecked != 1 {
				continue
			}
			body[value.Key] = value.Value
		}
		data, _ := json.Marshal(body)
		s = string(data)
	case UrlencodeMode:
		body := make(map[string]interface{})
		for _, value := range b.Parameter {
			if value.IsChecked != 1 {
				continue
			}
			body[value.Key] = value.Value
		}
		data, _ := json.Marshal(body)
		s = string(data)
	default:
		s = b.Raw
	}
	return
}

type Header struct {
	Parameter []*VarForm `json:"parameter" bson:"parameter"`
}

type Query struct {
	Parameter []*VarForm `json:"parameter" bson:"parameter"`
}

type Cookie struct {
	Parameter []*VarForm
}

type RegularExpression struct {
	Var     string `json:"var"`     // 变量
	Express string `json:"express"` // 表达式
	Val     string `json:"val"`     // 值
}

// Extract 提取response 中的值
func (re RegularExpression) Extract(str string, parameters *sync.Map) (value string) {
	name := tools.VariablesMatch(re.Var)
	if value = tools.FindDestStr(str, re.Express); value != "" {
		re.Val = value
		parameters.Store(name, value)
	}
	return
}

// VarForm 参数表
type VarForm struct {
	IsChecked   int64       `json:"is_checked" bson:"is_checked"`
	Type        string      `json:"type" bson:"type"`
	Key         string      `json:"key" bson:"key"`
	Value       interface{} `json:"value" bson:"value"`
	NotNull     int64       `json:"not_null" bson:"not_null"`
	Description string      `json:"description" bson:"description"`
	FieldType   string      `json:"field_type" bson:"field_type"`
}
type KV struct {
	Key   string `json:"key" bson:"key"`
	Value string `json:"value" bson:"value"`
}

type Bearer struct {
	Key string `json:"key" bson:"key"`
}

type Basic struct {
	UserName string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
}
type Auth struct {
	Type   string `json:"type" bson:"type"`
	KV     KV     `json:"kv" bson:"kv"`
	Bearer Bearer `json:"bearer" bson:"bearer"`
	Basic  Basic  `json:"basic" bson:"basic"`
}

// Conversion 将string转换为其他类型
func (v *VarForm) Conversion() {
	switch v.Type {
	case StringType:
		// 字符串类型不用转换
	case TextType:
		// 文本类型不用转换
	case ObjectType:
		// 对象不用转换
	case ArrayType:
		// 数组不用转换
	case IntegerType:
		value, err := strconv.ParseInt(v.Value.(string), 10, 64)
		if err == nil {
			v.Value = value
		}
	case NumberType:
		value, err := strconv.ParseInt(v.Value.(string), 10, 64)
		if err == nil {
			v.Value = value
		}
	case FloatType:
		value, err := strconv.ParseFloat(v.Value.(string), 32)
		if err == nil {
			v.Value = value
		}
	case DoubleType:
		value, err := strconv.ParseFloat(v.Value.(string), 64)
		if err == nil {
			v.Value = value
		}
	case FileType:
	case DateType:
	case DateTimeType:
	case TimeStampType:

	case BooleanType:
		if v.Value == "true" {
			v.Value = true
		}
		if v.Value == "false" {
			v.Value = false
		}
	}
}

// ReplaceQueryParameterizes 替换query中的变量
func (r *Api) ReplaceQueryParameterizes() {
	urls := tools.FindAllDestStr(r.Request.URL, "{{(.*?)}}")
	if urls != nil {
		for _, v := range urls {
			if value, ok := r.Parameters.Load(v[1]); ok {
				r.Request.URL = strings.Replace(r.Request.URL, v[0], value.(string), -1)
			}
		}
	}
	if r.Request.Body != nil {
		r.ReplaceBodyVarForm()
	}
	if r.Request.Query != nil {
		r.ReplaceQueryVarForm()
	}
	if r.Request.Header != nil {
		r.ReplaceHeaderVarForm()
	}
	if r.Request.Auth != nil {
		r.ReplaceAuthVarForm()
	}

}

func (r *Api) ReplaceCookieVarForm() {

}
func (r *Api) ReplaceBodyVarForm() {
	if r.Request.Body == nil {
		return
	}
	switch r.Request.Body.Mode {
	case NoneMode:
	case FormMode:
		if r.Request.Body.Parameter != nil {
			for _, parameter := range r.Request.Body.Parameter {
				keys := tools.FindAllDestStr(parameter.Key, "{{(.*?)}}")
				if keys != nil {
					for _, v := range keys {
						if value, ok := r.Parameters.Load(v[1]); ok {
							parameter.Key = strings.Replace(parameter.Key, v[0], value.(string), -1)
						}
					}
				}
				values := tools.FindAllDestStr(parameter.Value.(string), "{{(.*?)}}")
				if values != nil {
					for _, v := range values {
						if value, ok := r.Parameters.Load(v[1]); ok {
							parameter.Value = strings.Replace(parameter.Value.(string), v[0], value.(string), -1)
						}
					}
				}
			}
		}

	case UrlencodeMode:
		if r.Request.Body.Parameter != nil {
			for _, parameter := range r.Request.Body.Parameter {
				keys := tools.FindAllDestStr(parameter.Key, "{{(.*?)}}")
				if keys != nil {
					for _, v := range keys {
						if value, ok := r.Parameters.Load(v[1]); ok {
							parameter.Key = strings.Replace(parameter.Key, v[0], value.(string), -1)
						}
					}
				}
				values := tools.FindAllDestStr(parameter.Value.(string), "{{(.*?)}}")
				if values != nil {
					for _, v := range values {
						if value, ok := r.Parameters.Load(v[1]); ok {
							parameter.Value = strings.Replace(parameter.Value.(string), v[0], value.(string), -1)

						}
					}
				}
				parameter.Conversion()
			}
		}
	default:
		bosys := tools.FindAllDestStr(r.Request.Body.Raw, "{{(.*?)}}")
		if bosys != nil {
			for _, v := range bosys {
				if value, ok := r.Parameters.Load(v[1]); ok {
					r.Request.Body.Raw = strings.Replace(r.Request.Body.Raw, v[0], value.(string), -1)
				}
			}
		}
	}

}

func (r *Api) ReplaceHeaderVarForm() {
	if r.Request.Header != nil && r.Request.Header.Parameter != nil {
		for _, queryVarForm := range r.Request.Header.Parameter {
			queryParameterizes := tools.FindAllDestStr(queryVarForm.Key, "{{(.*?)}}")
			if queryParameterizes != nil {
				for _, v := range queryParameterizes {
					if value, ok := r.Parameters.Load(v[1]); ok {
						queryVarForm.Key = strings.Replace(queryVarForm.Key, v[0], value.(string), -1)
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
			queryVarForm.Conversion()
		}
	}
}

func (r *Api) ReplaceQueryVarForm() {
	if r.Request.Query != nil && r.Request.Query.Parameter != nil {
		for _, queryVarForm := range r.Request.Header.Parameter {
			queryParameterizes := tools.FindAllDestStr(queryVarForm.Key, "{{(.*?)}}")
			if queryParameterizes != nil {
				for _, v := range queryParameterizes {
					if value, ok := r.Parameters.Load(v[1]); ok {
						queryVarForm.Key = strings.Replace(queryVarForm.Key, v[0], value.(string), -1)
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
}

func (r *Api) ReplaceAuthVarForm() {
	if r.Request.Auth != nil {
		switch r.Request.Auth.Type {
		case KVType:

			if r.Request.Auth.KV.Key != "" {
				values := tools.FindAllDestStr(r.Request.Auth.KV.Value, "{{(.*?)}}")
				if values != nil {
					for _, value := range values {
						if v, ok := r.Parameters.Load(value[1]); ok {
							r.Request.Auth.KV.Value = strings.Replace(r.Request.Auth.KV.Value, value[0], v.(string), -1)
						}
					}
				}
			}

		case BearerType:
			if r.Request.Auth.Bearer.Key != "" {
				values := tools.FindAllDestStr(r.Request.Auth.Bearer.Key, "{{(.*?)}}")
				if values != nil {
					for _, value := range values {
						if v, ok := r.Parameters.Load(value[1]); ok {
							r.Request.Auth.Bearer.Key = strings.Replace(r.Request.Auth.Bearer.Key, value[0], v.(string), -1)
						}
					}
				}
			}
		case BasicType:
			if r.Request.Auth.Basic.UserName != "" {
				names := tools.FindAllDestStr(r.Request.Auth.Basic.UserName, "{{(.*?)}}")
				if names != nil {
					for _, value := range names {
						if v, ok := r.Parameters.Load(value[1]); ok {
							r.Request.Auth.Basic.UserName = strings.Replace(r.Request.Auth.Basic.UserName, value[0], v.(string), -1)
						}
					}
				}

			}

			if r.Request.Auth.Basic.Password != "" {
				passwords := tools.FindAllDestStr(r.Request.Auth.Basic.Password, "{{(.*?)}}")
				if passwords != nil {
					for _, value := range passwords {
						if v, ok := r.Parameters.Load(value[1]); ok {
							r.Request.Auth.Basic.Password = strings.Replace(r.Request.Auth.Basic.Password, value[0], v.(string), -1)
						}
					}
				}
			}
		}
	}
}

// FindParameterizes 将请求中的变量全部放到一个map中
func (r *Api) FindParameterizes() {
	if r.Parameters == nil {
		r.Parameters = new(sync.Map)
	}
	urls := tools.FindAllDestStr(r.Request.URL, "{{(.*?)}}")
	for _, name := range urls {
		if _, ok := r.Parameters.Load(name[1]); !ok {
			r.Parameters.Store(name[1], name[0])
		}
	}
	r.findBodyParameters()
	r.findQueryParameters()
	r.findHeaderParameters()
	r.findAuthParameters()
}

// ReplaceParameters 将场景变量中的值赋值给，接口变量
func (r *Api) ReplaceParameters(configuration *Configuration) {
	if r.Parameters == nil {
		r.Parameters = new(sync.Map)
	}

	r.Parameters.Range(func(k, v any) bool {
		if configuration.Variable != nil {
			if value, ok := configuration.Variable.Load(k); ok {
				if v == fmt.Sprintf("{{%s}}", k) {
					r.Parameters.Store(k, value)
				}
			}
		}
		if configuration.ParameterizedFile != nil {
			if _, ok := configuration.ParameterizedFile.VariableNames.VarMapList[k.(string)]; ok {
				if v == fmt.Sprintf("{{%s}}", k) {
					configuration.Mu.Lock()
					value := configuration.ParameterizedFile.UseVar(k.(string))
					r.Parameters.Store(k, value)
					configuration.Mu.Unlock()
				}
			}
		}
		return true
	})
}

// 将Query中的变量，都存储到接口变量中
func (r *Api) findQueryParameters() {

	if r.Request.Query == nil || r.Request.Query.Parameter == nil {
		return
	}
	for _, varForm := range r.Request.Query.Parameter {
		nameParameters := tools.FindAllDestStr(varForm.Key, "{{(.*?)}}")
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

func (r *Api) findBodyParameters() {
	if r.Request.Body != nil {
		switch r.Request.Body.Mode {
		case NoneMode:
		case FormMode:
			if r.Request.Body.Parameter == nil {
				return
			}
			for _, parameter := range r.Request.Body.Parameter {
				keys := tools.FindAllDestStr(parameter.Key, "{{(.*?)}}")
				if keys != nil && len(keys) > 1 {
					for _, key := range keys {
						if _, ok := r.Parameters.Load(key[1]); !ok {
							r.Parameters.Store(key[1], key[0])
						}
					}
				}
				values := tools.FindAllDestStr(parameter.Value.(string), "{{(.*?)}}")
				if values != nil {
					for _, value := range values {
						if _, ok := r.Parameters.Load(value[1]); !ok {
							r.Parameters.Store(value[1], value[0])
						}
					}
				}

			}
		case UrlencodeMode:
			if r.Request.Body.Parameter == nil {
				return
			}
			for _, parameter := range r.Request.Body.Parameter {
				keys := tools.FindAllDestStr(parameter.Key, "{{(.*?)}}")
				if keys != nil {
					for _, key := range keys {
						if _, ok := r.Parameters.Load(key[1]); !ok {
							r.Parameters.Store(key[1], key[0])
						}
					}
				}
				values := tools.FindAllDestStr(parameter.Value.(string), "{{(.*?)}}")
				if values != nil {
					for _, value := range values {
						if _, ok := r.Parameters.Load(value[1]); !ok {
							r.Parameters.Store(value[1], value[0])
						}
					}
				}
			}
		default:
			if r.Request.Body.Raw == "" {
				return
			}
			bodys := tools.FindAllDestStr(r.Request.Body.Raw, "{{(.*?)}}")
			if bodys != nil {
				for _, body := range bodys {
					if len(body) > 1 {
						if _, ok := r.Parameters.Load(body[1]); !ok {
							r.Parameters.Store(body[1], body[0])
						}
					}
				}
			}
		}
	}

}

// 将Header中的变量，都存储到接口变量中
func (r *Api) findHeaderParameters() {

	if r.Request.Header == nil || r.Request.Header.Parameter == nil {
		return
	}
	for _, varForm := range r.Request.Header.Parameter {
		nameParameters := tools.FindAllDestStr(varForm.Key, "{{(.*?)}}")
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

func (r *Api) findAuthParameters() {
	if r.Request.Auth != nil {
		switch r.Request.Auth.Type {
		case KVType:
			if r.Request.Auth.KV.Key == "" {
				return
			}
			keys := tools.FindAllDestStr(r.Request.Auth.KV.Key, "{{(.*?)}}")
			for _, key := range keys {
				if _, ok := r.Parameters.Load(key[1]); !ok {
					r.Parameters.Store(key[1], key[0])
				}
			}
			if r.Request.Auth.KV.Value == "" {
				return
			}
			values := tools.FindAllDestStr(r.Request.Auth.KV.Value, "{{(.*?)}}")
			for _, value := range values {
				if _, ok := r.Parameters.Load(value[1]); !ok {
					r.Parameters.Store(value[1], value[0])
				}
			}
		case BearerType:
			if r.Request.Auth.Bearer.Key == "" {
				return
			}
			keys := tools.FindAllDestStr(r.Request.Auth.Bearer.Key, "{{(.*?)}}")
			for _, key := range keys {
				if _, ok := r.Parameters.Load(key[1]); !ok {
					r.Parameters.Store(key[1], key[0])
				}
			}
		case BasicType:
			if r.Request.Auth.Basic.UserName == "" {
				return
			}
			names := tools.FindAllDestStr(r.Request.Auth.Basic.UserName, "{{(.*?)}}")
			for _, name := range names {
				if _, ok := r.Parameters.Load(name[1]); !ok {
					r.Parameters.Store(name[1], name[0])
				}
			}
			if r.Request.Auth.Basic.UserName == "" {
				return
			}
			pws := tools.FindAllDestStr(r.Request.Auth.Basic.Password, "{{(.*?)}}")
			for _, pw := range pws {
				if _, ok := r.Parameters.Load(pw[1]); !ok {
					r.Parameters.Store(pw[1], pw[0])
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
