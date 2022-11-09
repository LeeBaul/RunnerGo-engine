package model

import (
	"RunnerGo-engine/log"
	"RunnerGo-engine/tools"
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"github.com/ThomsonReutersEikon/go-ntlm/ntlm"
	"github.com/comcast/go-edgegrid/edgegrid"
	hawk "github.com/hiyosi/hawk"
	"github.com/lixiangyun/go-ntlm/messages"
	uuid "github.com/satori/go.uuid"
	"github.com/valyala/fasthttp"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Api 请求数据
type Api struct {
	TargetId   int64                `json:"target_id" bson:"target_id"`
	Uuid       uuid.UUID            `json:"uuid" bson:"uuid"`
	Name       string               `json:"name" bson:"name"`
	TeamId     int64                `json:"team_id" bson:"team_id"`
	TargetType string               `json:"target_type" bson:"target_type"` // api/webSocket/tcp/grpc
	Method     string               `json:"method" bson:"method"`           // 方法 GET/POST/PUT
	Request    Request              `json:"request" bson:"request"`
	Parameters *sync.Map            `json:"parameters" bson:"parameters"`
	Assert     []*AssertionText     `json:"assert" bson:"assert"`         // 验证的方法(断言)
	Timeout    int64                `json:"timeout" bson:"timeout"`       // 请求超时时间
	Regex      []*RegularExpression `json:"regex" bson:"regex"`           // 正则表达式
	Debug      string               `json:"debug" bson:"debug"`           // 是否开启Debug模式
	Connection int64                `json:"connection" bson:"connection"` // 0:websocket长连接
	Variable   []*KV                `json:"variable" bson:"variable"`     // 全局变量
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

func (b *Body) SetBody(req *fasthttp.Request) string {
	if b == nil {
		return ""
	}
	switch b.Mode {
	case NoneMode:
	case FormMode:
		req.Header.SetContentType("multipart/form-data")
		// 新建一个缓冲，用于存放文件内容

		if b.Parameter == nil {
			b.Parameter = []*VarForm{}
		}

		bodyBuffer := &bytes.Buffer{}
		bodyWriter := multipart.NewWriter(bodyBuffer)

		contentType := bodyWriter.FormDataContentType()
		for _, value := range b.Parameter {

			if value.IsChecked != 1 {
				continue
			}
			if value.Key == "" {
				continue
			}

			if value.Type == FileType {
				if value.FileBase64 == nil || len(value.FileBase64) < 1 {
					continue
				}
				for _, base64Str := range value.FileBase64 {
					by := tools.Base64DeEncode(base64Str, FileType)
					if by == nil {
						continue
					}
					fileWriter, err := bodyWriter.CreateFormFile(value.Key, value.Value.(string))
					//fileType := strings.Split(value.Value.(string), ".")[1]
					if err != nil {
						log.Logger.Error("CreateFormFile失败： ", err)
						continue
					}

					file := bytes.NewReader(by)
					_, err = io.Copy(fileWriter, file)
					if err != nil {
						continue
					}
				}
			} else {
				filedWriter, err := bodyWriter.CreateFormField(value.Key)
				by := value.toByte()
				filed := bytes.NewReader(by)
				_, err = io.Copy(filedWriter, filed)
				if err != nil {
					log.Logger.Error("CreateFormFile失败： ", err)
					continue
				}
			}

		}
		bodyWriter.Close()
		req.Header.SetContentType(contentType)
		if bodyBuffer.Bytes() != nil && bodyBuffer.Len() != 68 {
			req.SetBody(bodyBuffer.Bytes())
		}
		return bodyBuffer.String()
	case UrlencodeMode:
		req.Header.SetContentType("application/x-www-form-urlencoded")
		args := req.PostArgs()
		for _, value := range b.Parameter {
			if value.IsChecked != 1 || value.Key == "" || value.Value == nil {
				continue
			}
			args.Add(value.Key, value.Value.(string))

		}
		req.SetBodyString(args.String())
		return args.String()
	case XmlMode:
		req.Header.SetContentType("application/xml")
		req.SetBodyString(b.Raw)
		return b.Raw
	case JSMode:
		req.Header.SetContentType("application/javascript")
		req.SetBodyString(b.Raw)
		return b.Raw
	case PlainMode:
		req.Header.SetContentType("text/plain")
		req.SetBodyString(b.Raw)
		return b.Raw
	case HtmlMode:
		req.Header.SetContentType("text/html")
		req.SetBodyString(b.Raw)
		return b.Raw
	case JsonMode:
		req.Header.SetContentType("application/json")
		req.SetBodyString(b.Raw)
		return b.Raw
	}
	return ""
}

type Header struct {
	Parameter []*VarForm `json:"parameter" bson:"parameter"`
}

func (header *Header) SetHeader(req *fasthttp.Request) {
	if header != nil && header.Parameter != nil {
		for _, v := range header.Parameter {
			if v.IsChecked == 1 {
				if v.Value == nil {
					continue
				}
				if strings.EqualFold(v.Key, "content-type") {
					req.Header.SetContentType(v.Value.(string))
				}
				if strings.EqualFold(v.Key, "host") {
					req.Header.SetHost(v.Value.(string))
				}
				req.Header.Set(v.Key, v.Value.(string))
			}
		}

	}
}

type Query struct {
	Parameter []*VarForm `json:"parameter" bson:"parameter"`
}

type Cookie struct {
	Parameter []*VarForm
}

type RegularExpression struct {
	IsChecked int         `json:"is_checked"` // 1 选中, -1未选
	Type      int         `json:"type"`       // 0 正则  1 json
	Var       string      `json:"var"`        // 变量
	Express   string      `json:"express"`    // 表达式
	Val       interface{} `json:"val"`        // 值
}

// Extract 提取response 中的值
func (re RegularExpression) Extract(str string, configuration *Configuration) (value interface{}) {
	name := tools.VariablesMatch(re.Var)
	switch re.Type {
	case 0:
		kv := &KV{
			Key: name,
		}
		if value = tools.FindDestStr(str, re.Express); value != "" {
			re.Val = value
			kv.Value = value
		}
		configuration.Variable = append(configuration.Variable, kv)
	case 1:
		kv := &KV{
			Key: name,
		}
		value = tools.JsonPath(str, re.Express)
		kv.Value = value
		configuration.Variable = append(configuration.Variable, kv)
	}

	return
}

// VarForm 参数表
type VarForm struct {
	IsChecked   int64       `json:"is_checked" bson:"is_checked"`
	Type        string      `json:"type" bson:"type"`
	FileBase64  []string    `json:"fileBase64"`
	Key         string      `json:"key" bson:"key"`
	Value       interface{} `json:"value" bson:"value"`
	NotNull     int64       `json:"not_null" bson:"not_null"`
	Description string      `json:"description" bson:"description"`
	FieldType   string      `json:"field_type" bson:"field_type"`
}

type KV struct {
	Key   string      `json:"key" bson:"key"`
	Value interface{} `json:"value" bson:"value"`
}

type PlanKv struct {
	Var string `json:"Var"`
	Val string `json:"Val"`
}

type Form struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type Bearer struct {
	Key string `json:"key" bson:"key"`
}

type Basic struct {
	UserName string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
}

type Digest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Realm     string `json:"realm"`
	Nonce     string `json:"nonce"`
	Algorithm string `json:"algorithm"`
	Qop       string `json:"qop"`
	Nc        string `json:"nc"`
	Cnonce    string `json:"cnonce"`
	Opaque    string `json:"opaque"`
}

type Hawk struct {
	AuthID             string `json:"authId"`
	AuthKey            string `json:"authKey"`
	Algorithm          string `json:"algorithm"`
	User               string `json:"user"`
	Nonce              string `json:"nonce"`
	ExtraData          string `json:"extraData"`
	App                string `json:"app"`
	Delegation         string `json:"delegation"`
	Timestamp          string `json:"timestamp"`
	IncludePayloadHash int    `json:"includePayloadHash"`
}

type AwsV4 struct {
	AccessKey          string `json:"accessKey"`
	SecretKey          string `json:"secretKey"`
	Region             string `json:"region"`
	Service            string `json:"service"`
	SessionToken       string `json:"sessionToken"`
	AddAuthDataToQuery int    `json:"addAuthDataToQuery"`
}

type Ntlm struct {
	Username            string `json:"username"`
	Password            string `json:"password"`
	Domain              string `json:"domain"`
	Workstation         string `json:"workstation"`
	DisableRetryRequest int    `json:"disableRetryRequest"`
}

type Edgegrid struct {
	AccessToken   string `json:"accessToken"`
	ClientToken   string `json:"clientToken"`
	ClientSecret  string `json:"clientSecret"`
	Nonce         string `json:"nonce"`
	Timestamp     string `json:"timestamp"`
	BaseURi       string `json:"baseURi"`
	HeadersToSign string `json:"headersToSign"`
}

type Oauth1 struct {
	ConsumerKey          string `json:"consumerKey"`
	ConsumerSecret       string `json:"consumerSecret"`
	SignatureMethod      string `json:"signatureMethod"`
	AddEmptyParamsToSign int    `json:"addEmptyParamsToSign"`
	IncludeBodyHash      int    `json:"includeBodyHash"`
	AddParamsToHeader    int    `json:"addParamsToHeader"`
	Realm                string `json:"realm"`
	Version              string `json:"version"`
	Nonce                string `json:"nonce"`
	Timestamp            string `json:"timestamp"`
	Verifier             string `json:"verifier"`
	Callback             string `json:"callback"`
	TokenSecret          string `json:"tokenSecret"`
	Token                string `json:"token"`
}
type Auth struct {
	Type     string    `json:"type" bson:"type"`
	KV       *KV       `json:"kv" bson:"kv"`
	Bearer   *Bearer   `json:"bearer" bson:"bearer"`
	Basic    *Basic    `json:"basic" bson:"basic"`
	Digest   *Digest   `json:"digest"`
	Hawk     *Hawk     `json:"hawk"`
	Awsv4    *AwsV4    `json:"awsv4"`
	Ntlm     *Ntlm     `json:"ntlm"`
	Edgegrid *Edgegrid `json:"edgegrid"`
	Oauth1   *Oauth1   `json:"oauth1"`
}

type Token struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

type RequestData struct {
	Url           string `json:"url"`
	Method        string `json:"method"`
	Data          string `json:"data"`
	OauthCallback string `json:"oauth_callback"`
}

type Consumer struct {
}

func (auth *Auth) SetAuth(req *fasthttp.Request) {
	if auth != nil && auth.Type != NoAuth {
		switch auth.Type {
		case Kv:
			if auth.KV.Value != nil {
				req.Header.Add(auth.KV.Key, auth.KV.Value.(string))
			}

		case BEarer:
			req.Header.Add("authorization", "Bearer "+auth.Bearer.Key)
		case BAsic:
			req.Header.Add("authorization", "Basic "+string(tools.Base64Encode(auth.Basic.UserName+auth.Basic.Password)))
		case DigestType:
			encryption := tools.GetEncryption(auth.Digest.Algorithm)
			if encryption != nil {
				uri := string(req.URI().RequestURI())
				ha1 := ""
				ha2 := ""
				response := ""
				if auth.Digest.Cnonce == "" {
					auth.Digest.Cnonce = "apipost"
				}
				if auth.Digest.Nc == "" {
					auth.Digest.Nc = "00000001"
				}
				if strings.HasSuffix(auth.Digest.Algorithm, "-sess") {
					ha1 = encryption.HashFunc(encryption.HashFunc(auth.Digest.Username+":"+auth.Digest.Realm+":"+
						auth.Digest.Password) + ":" + auth.Digest.Nonce + ":" + auth.Digest.Cnonce)
				} else {
					ha1 = encryption.HashFunc(auth.Digest.Username + ":" + auth.Digest.Realm + ":" + auth.Digest.Password)
				}
				if auth.Digest.Qop != "auth-int" {
					ha2 = encryption.HashFunc(string(req.Header.Method()) + req.URI().String())
				} else {
					ha2 = encryption.HashFunc(string(req.Header.Method()) + uri + encryption.HashFunc(string(req.Body())))
				}
				if auth.Digest.Qop == "auth" || auth.Digest.Qop == "authn-int" {
					response = encryption.HashFunc(ha1 + ":" + auth.Digest.Nonce + ":" + auth.Digest.Nc +
						auth.Digest.Cnonce + ":" + auth.Digest.Qop + ":" + ha2)
				} else {
					response = encryption.HashFunc(ha1 + ":" + auth.Digest.Nonce + ":" + ha2)
				}
				digest := fmt.Sprintf("username=%s, realm=%s, nonce=%s, uri=%s, algorithm=%s, qop=%s, nc=%s, cnonce=%s, response=%s, opaque=%s",
					auth.Digest.Username, auth.Digest.Realm, auth.Digest.Nonce, uri, auth.Digest.Algorithm, auth.Digest.Qop,
					auth.Digest.Nc, auth.Digest.Cnonce, response, auth.Digest.Opaque)
				req.Header.Add("Authorization", digest)
			}
		case HawkType:
			var alg hawk.Alg
			if strings.Contains(auth.Hawk.Algorithm, "SHA512") {
				alg = 2
			} else {
				alg = 1
			}
			credential := &hawk.Credential{
				ID:  auth.Hawk.AuthID,
				Key: auth.Hawk.AuthKey,
				Alg: alg,
			}
			timestamp, err := strconv.ParseInt(auth.Hawk.Timestamp, 10, 64)
			if err != nil {
				timestamp = time.Now().Unix()
			}
			option := &hawk.Option{
				TimeStamp: timestamp,
				Nonce:     auth.Hawk.Nonce,
				Ext:       auth.Hawk.ExtraData,
			}
			c := hawk.NewClient(credential, option)
			authorization, _ := c.Header(string(req.Header.Method()), string(req.Host())+string(req.Header.RequestURI()))
			req.Header.Add("Authorization", authorization)
		case EdgegridType:
			reader := bytes.NewReader(req.Body())
			reqNew, err := http.NewRequest(string(req.Header.Method()), req.URI().String(), reader)
			if err != nil {
				return
			}
			params := edgegrid.NewAuthParams(reqNew, auth.Edgegrid.AccessToken, auth.Edgegrid.ClientToken, auth.Edgegrid.ClientSecret)
			authorization := edgegrid.Auth(params)
			req.Header.Add("Authorization", authorization)
		case NtlmType:
			session, err := ntlm.CreateClientSession(ntlm.Version1, ntlm.ConnectionlessMode)
			if err != nil {
				return
			}
			session.SetUserInfo(auth.Ntlm.Username, auth.Ntlm.Password, auth.Ntlm.Domain)
			negotiate, err := session.GenerateNegotiateMessage()
			if err != nil {
				return
			}
			challenge, err := messages.ParseAuthenticateMessage(negotiate.Bytes, 2)
			if err != nil {
				return
			}
			req.Header.Add("Connection", "keep-alive")
			req.Header.Add("Authorization", challenge.String())
		case Awsv4Type:
			signature := ""
			date := strconv.Itoa(int(time.Now().Month())) + strconv.Itoa(time.Now().Day())
			awsv := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/2022%s/%s/%s/aws4_request, SignedHeaders=content-length;content-type;host;x-amz-date;x-amz-security-token, Signature=%s",
				auth.Awsv4.AccessKey, date,
				auth.Awsv4.Region, auth.Awsv4.Service, signature)
			currentTime := strconv.Itoa(time.Now().Hour()) + strconv.Itoa(time.Now().Minute()) + strconv.Itoa(time.Now().Second())
			req.Header.Add("X-Amz-Security-Token", auth.Awsv4.SessionToken)
			req.Header.Add("X-Amz-Date", date+"T"+currentTime+"Z")
			req.Header.Add("Authorization", awsv)

		case Oauth1Type:

			//config := oauth1.Config{
			//	ConsumerKey:    auth.Oauth1.ConsumerKey,
			//	ConsumerSecret: auth.Oauth1.ConsumerSecret,
			//	CallbackURL:    req.URI().String(),
			//	Endpoint:       twitter.AuthorizeEndpoint,
			//	Realm:          auth.Oauth1.Realm,
			//}
			//
			//token := oauth1.NewToken(auth.Oauth1.Token, auth.Oauth1.Callback)
			//
			//authorization :=
			//	req.Header.Add("Authorization", authorization)
		}
	}
}

func (v *VarForm) toByte() (by []byte) {
	if v.Value == nil {
		return
	}
	switch v.FieldType {
	case StringType:
		by = []byte(v.Value.(string))
	case TextType:
		by = []byte(v.Value.(string))
	case ObjectType:
		by = []byte(v.Value.(string))
	case ArrayType:
		by = []byte(v.Value.(string))
	case NumberType:
		bytesBuffer := bytes.NewBuffer([]byte{})
		_ = binary.Write(bytesBuffer, binary.BigEndian, v.Value.(int))
		by = bytesBuffer.Bytes()
	case IntegerType:
		bytesBuffer := bytes.NewBuffer([]byte{})
		_ = binary.Write(bytesBuffer, binary.BigEndian, v.Value.(int))
		by = bytesBuffer.Bytes()
	case DoubleType:
		bytesBuffer := bytes.NewBuffer([]byte{})
		_ = binary.Write(bytesBuffer, binary.BigEndian, v.Value.(int64))
		by = bytesBuffer.Bytes()
	case FileType:
		bits := math.Float64bits(v.Value.(float64))
		binary.LittleEndian.PutUint64(by, bits)
	case BooleanType:
		buf := bytes.Buffer{}
		enc := gob.NewEncoder(&buf)
		_ = enc.Encode(v.Value.(bool))
		by = buf.Bytes()
	case DateType:
		by = []byte(v.Value.(string))
	case DateTimeType:
		by = []byte(v.Value.(string))
	case TimeStampType:
		bytesBuffer := bytes.NewBuffer([]byte{})
		_ = binary.Write(bytesBuffer, binary.BigEndian, v.Value.(int64))
		by = bytesBuffer.Bytes()

	}
	return
}

// Conversion 将string转换为其他类型
func (v *VarForm) Conversion() {
	if v.Value == nil {
		return
	}
	switch v.FieldType {
	case StringType:
		v.Value = v.Value.(string)
		// 字符串类型不用转换
	case TextType:
		v.Value = v.Value.(string)
		// 文本类型不用转换
	case ObjectType:
		v.Value = v.Value.(string)
		// 对象不用转换
	case ArrayType:
		v.Value = v.Value.(string)
		// 数组不用转换
	case IntegerType:
		v.Value = v.Value.(int)
	case NumberType:
		v.Value = v.Value.(int)
	case FloatType:
		v.Value = v.Value.(float64)
	case DoubleType:
		v.Value = v.Value.(float64)
	case FileType:
		v.Value = v.Value.(string)
	case DateType:
		v.Value = v.Value.(string)
	case DateTimeType:
		v.Value = v.Value.(string)
	case TimeStampType:
		v.Value = v.Value.(int64)
	case BooleanType:
		v.Value = v.Value.(bool)
	}
}

// ReplaceQueryParameterizes 替换query中的变量
func (r *Api) ReplaceQueryParameterizes() {
	urls := tools.FindAllDestStr(r.Request.URL, "{{(.*?)}}")
	if urls != nil {
		for _, v := range urls {
			if r.Parameters != nil {
				r.Parameters.Range(func(key, value any) bool {
					return true
				})
				if value, ok := r.Parameters.Load(v[1]); ok {
					if value == nil {
						continue
					}
					r.Request.URL = strings.Replace(r.Request.URL, v[0], value.(string), -1)
				}
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
		if r.Request.Body.Parameter != nil && len(r.Request.Body.Parameter) > 0 {
			for _, parameter := range r.Request.Body.Parameter {
				keys := tools.FindAllDestStr(parameter.Key, "{{(.*?)}}")
				if keys != nil {
					for _, v := range keys {
						if value, ok := r.Parameters.Load(v[1]); ok {
							if value == nil {
								continue
							}
							parameter.Key = strings.Replace(parameter.Key, v[0], value.(string), -1)
						}
					}
				}
				if parameter.Value == nil {
					continue
				}
				values := tools.FindAllDestStr(parameter.Value.(string), "{{(.*?)}}")
				if values != nil {
					for _, v := range values {
						if value, ok := r.Parameters.Load(v[1]); ok {
							if value == nil || parameter.Value == nil {
								continue
							}
							parameter.Value = strings.Replace(parameter.Value.(string), v[0], value.(string), -1)
						}
					}
				}
			}
		}

	case UrlencodeMode:
		if r.Request.Body.Parameter != nil && len(r.Request.Body.Parameter) > 0 {
			for _, parameter := range r.Request.Body.Parameter {
				keys := tools.FindAllDestStr(parameter.Key, "{{(.*?)}}")
				if keys != nil {
					for _, v := range keys {
						if value, ok := r.Parameters.Load(v[1]); ok {
							if value == nil {
								continue
							}
							parameter.Key = strings.Replace(parameter.Key, v[0], value.(string), -1)
						}
					}
				}
				if parameter.Value == nil {
					continue
				}
				values := tools.FindAllDestStr(parameter.Value.(string), "{{(.*?)}}")
				if values != nil {
					for _, v := range values {
						if value, ok := r.Parameters.Load(v[1]); ok {
							if value == nil || parameter.Value == nil {
								continue
							}
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
					if value == nil {
						continue
					}
					r.Request.Body.Raw = strings.Replace(r.Request.Body.Raw, v[0], value.(string), -1)
				}
			}
		}
	}

}

func (r *Api) ReplaceHeaderVarForm() {
	if r.Request.Header != nil && r.Request.Header.Parameter != nil && len(r.Request.Header.Parameter) > 0 {
		for _, queryVarForm := range r.Request.Header.Parameter {
			queryParameterizes := tools.FindAllDestStr(queryVarForm.Key, "{{(.*?)}}")
			if queryParameterizes != nil {
				for _, v := range queryParameterizes {
					if value, ok := r.Parameters.Load(v[1]); ok {
						if value == nil {
							continue
						}
						queryVarForm.Key = strings.Replace(queryVarForm.Key, v[0], value.(string), -1)
					}
				}
			}
			if queryVarForm.Value == nil {
				continue
			}
			queryParameterizes = tools.FindAllDestStr(queryVarForm.Value.(string), "{{(.*?)}}")
			if queryParameterizes != nil {
				for _, v := range queryParameterizes {
					if value, ok := r.Parameters.Load(v[1]); ok {
						if queryVarForm.Value == nil || value == nil {
							continue
						}
						queryVarForm.Value = strings.Replace(queryVarForm.Value.(string), v[0], value.(string), -1)
					}
				}
			}
			queryVarForm.Conversion()
		}
	}
}

func (r *Api) ReplaceQueryVarForm() {
	if r.Request.Query != nil && r.Request.Query.Parameter != nil && len(r.Request.Query.Parameter) > 0 {
		if r.Request.Header.Parameter != nil && len(r.Request.Header.Parameter) > 0 {
			for _, queryVarForm := range r.Request.Header.Parameter {
				if queryVarForm.Value == nil {
					continue
				}
				queryParameterizes := tools.FindAllDestStr(queryVarForm.Key, "{{(.*?)}}")
				if queryParameterizes != nil {
					for _, v := range queryParameterizes {
						if value, ok := r.Parameters.Load(v[1]); ok {
							if value == nil {
								continue
							}
							queryVarForm.Key = strings.Replace(queryVarForm.Key, v[0], value.(string), -1)
						}
					}
				}
				queryParameterizes = tools.FindAllDestStr(queryVarForm.Value.(string), "{{(.*?)}}")
				if queryParameterizes != nil {
					for _, v := range queryParameterizes {
						if value, ok := r.Parameters.Load(v[1]); ok {
							if value == nil {
								continue
							}
							queryVarForm.Value = strings.Replace(queryVarForm.Value.(string), v[0], value.(string), -1)
						}
					}
				}
			}
		}

	}
}

func (r *Api) ReplaceAuthVarForm() {

	if r.Request.Auth != nil && r.Request.Auth.KV.Value != nil {
		switch r.Request.Auth.Type {
		case Kv:

			if r.Request.Auth.KV != nil && r.Request.Auth.KV.Key != "" {
				values := tools.FindAllDestStr(r.Request.Auth.KV.Value.(string), "{{(.*?)}}")
				if values != nil {
					for _, value := range values {
						if v, ok := r.Parameters.Load(value[1]); ok {
							if v != nil {
								r.Request.Auth.KV.Value = strings.Replace(r.Request.Auth.KV.Value.(string), value[0], v.(string), -1)
							}

						}
					}
				}
			}

		case BEarer:
			if r.Request.Auth.Bearer != nil && r.Request.Auth.Bearer.Key != "" {
				values := tools.FindAllDestStr(r.Request.Auth.Bearer.Key, "{{(.*?)}}")
				if values != nil {
					for _, value := range values {
						if v, ok := r.Parameters.Load(value[1]); ok {
							if v != nil {
								r.Request.Auth.Bearer.Key = strings.Replace(r.Request.Auth.Bearer.Key, value[0], v.(string), -1)
							}
						}
					}
				}
			}
		case BAsic:
			if r.Request.Auth.Basic != nil && r.Request.Auth.Basic.UserName != "" {
				names := tools.FindAllDestStr(r.Request.Auth.Basic.UserName, "{{(.*?)}}")
				if names != nil {
					for _, value := range names {
						if v, ok := r.Parameters.Load(value[1]); ok {
							if v != nil {
								r.Request.Auth.Basic.UserName = strings.Replace(r.Request.Auth.Basic.UserName, value[0], v.(string), -1)
							}
						}
					}
				}

			}

			if r.Request.Auth.Basic != nil && r.Request.Auth.Basic.Password != "" {
				passwords := tools.FindAllDestStr(r.Request.Auth.Basic.Password, "{{(.*?)}}")
				if passwords != nil {
					for _, value := range passwords {
						if v, ok := r.Parameters.Load(value[1]); ok {
							if v != nil {
								r.Request.Auth.Basic.Password = strings.Replace(r.Request.Auth.Basic.Password, value[0], v.(string), -1)
							}
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
	r.Request.URL = strings.TrimSpace(r.Request.URL)
	urls := tools.FindAllDestStr(r.Request.URL, "{{(.*?)}}")

	for _, name := range urls {

		r.Parameters.Range(func(key, value any) bool {
			return true
		})
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
		if k == nil {
			return true
		}

		if configuration.Variable != nil {
			for _, kv := range configuration.Variable {
				if kv.Key == k {
					if v == fmt.Sprintf("{{%s}}", k) {
						r.Parameters.Store(k, kv.Value)
					}
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
		if varForm.Value == nil {
			continue
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
				if parameter.Value == nil {
					continue
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
				if parameter.Value == nil {
					continue
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
		if varForm.Value == nil {
			continue
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
		case Kv:
			if r.Request.Auth.KV.Key == "" {
				return
			}
			keys := tools.FindAllDestStr(r.Request.Auth.KV.Key, "{{(.*?)}}")
			for _, key := range keys {
				if _, ok := r.Parameters.Load(key[1]); !ok {
					r.Parameters.Store(key[1], key[0])
				}
			}

			if r.Request.Auth.KV.Value == nil {
				return

			}

			values := tools.FindAllDestStr(r.Request.Auth.KV.Value.(string), "{{(.*?)}}")
			for _, value := range values {
				if _, ok := r.Parameters.Load(value[1]); !ok {
					r.Parameters.Store(value[1], value[0])
				}
			}
		case BEarer:
			if r.Request.Auth.Bearer.Key == "" {
				return
			}
			keys := tools.FindAllDestStr(r.Request.Auth.Bearer.Key, "{{(.*?)}}")
			for _, key := range keys {
				if _, ok := r.Parameters.Load(key[1]); !ok {
					r.Parameters.Store(key[1], key[0])
				}
			}
		case BAsic:
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
