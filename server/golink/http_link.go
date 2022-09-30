// Package golink 连接
package golink

import (
	"github.com/shopspring/decimal"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/mongo"
	log2 "kp-runner/log"
	"kp-runner/model"
	"kp-runner/server/client"
)

// HttpSend 发送http请求
func HttpSend(event model.Event, api model.Api, configuration *model.Configuration, requestCollection *mongo.Collection) (bool, int64, uint64, float64, float64, string, int64) {
	var (
		isSucceed     = true
		errCode       = model.NoError
		receivedBytes = float64(0)
		errMsg        = ""
		timestamp     = int64(0)
	)

	resp, req, requestTime, sendBytes, err, timestamp, str := client.HTTPRequest(api.Method, api.Request.URL, api.Request.Body, api.Request.Query,
		api.Request.Header, api.Request.Auth, api.Timeout)
	defer fasthttp.ReleaseResponse(resp) // 用完需要释放资源
	defer fasthttp.ReleaseRequest(req)
	var regex []map[string]interface{}
	if api.Regex != nil {
		for _, regular := range api.Regex {
			reg := make(map[string]interface{})
			value := regular.Extract(string(resp.Body()), configuration)
			reg[regular.Var] = value
			regex = append(regex, reg)
		}
	}
	if err != nil {
		isSucceed = false
		errMsg = err.Error()
	}
	var assertionMsgList []model.AssertionMsg
	// 断言验证
	if api.Assert != nil {
		var assertionMsg = model.AssertionMsg{}
		var (
			code    = int64(10000)
			succeed = true
			msg     = ""
		)
		for _, v := range api.Assert {
			code, succeed, msg = v.VerifyAssertionText(resp)
			if succeed != true {
				errCode = code
				isSucceed = succeed
				errMsg = msg
			}
			assertionMsg.Code = code
			assertionMsg.IsSucceed = succeed
			assertionMsg.Msg = msg
			assertionMsgList = append(assertionMsgList, assertionMsg)
		}
	}
	// 接收到的字节长度
	//contentLength = uint(resp.Header.ContentLength())

	receivedBytes, _ = decimal.NewFromFloat(float64(resp.Header.ContentLength()) / 1024).Round(2).Float64()
	// 开启debug模式后，将请求响应信息写入到mongodb中
	if api.Debug != "" && api.Debug != "stop" {
		switch api.Debug {
		case model.All:
			debugMsg := make(map[string]interface{})
			debugMsg["uuid"] = api.Uuid.String()
			debugMsg["event_id"] = event.Id
			debugMsg["api_id"] = api.TargetId
			debugMsg["api_name"] = api.Name
			debugMsg["type"] = model.RequestType
			debugMsg["request_time"] = requestTime / 1000000
			debugMsg["request_code"] = resp.StatusCode()
			debugMsg["request_header"] = req.Header.String()
			if string(req.Body()) != "" {
				debugMsg["request_body"] = string(req.Body())
			} else {
				debugMsg["request_body"] = str
			}

			debugMsg["response_header"] = resp.Header.String()
			debugMsg["response_body"] = string(resp.Body())
			debugMsg["response_bytes"] = receivedBytes
			if err != nil {
				debugMsg["response_body"] = err.Error()
			}
			switch isSucceed {
			case false:
				debugMsg["status"] = model.Failed
			case true:
				debugMsg["status"] = model.Success
			}

			debugMsg["next_list"] = event.NextList

			if api.Assert != nil {
				debugMsg["assertion"] = assertionMsgList
			}
			if api.Regex != nil {
				debugMsg["regex"] = regex
			}
			if requestCollection != nil {
				debugMsg["report_id"] = event.ReportId
				model.Insert(requestCollection, debugMsg)
			}
		case model.OnlySuccess:
			if isSucceed == true {
				debugMsg := make(map[string]interface{})
				debugMsg["uuid"] = api.Uuid.String()
				debugMsg["event_id"] = event.Id
				debugMsg["api_id"] = api.TargetId
				debugMsg["api_name"] = api.Name
				debugMsg["type"] = model.RequestType
				debugMsg["request_time"] = requestTime / 1000000
				debugMsg["request_code"] = resp.StatusCode()
				debugMsg["request_header"] = req.Header.String()
				if string(req.Body()) != "" {
					debugMsg["request_body"] = string(req.Body())
				} else {
					debugMsg["request_body"] = str
				}
				debugMsg["response_header"] = resp.Header.String()
				debugMsg["response_body"] = string(resp.Body())
				debugMsg["response_bytes"] = receivedBytes
				debugMsg["status"] = model.Success
				debugMsg["next_list"] = event.NextList
				if err != nil {
					debugMsg["response_body"] = err.Error()
				}
				if api.Assert != nil {
					debugMsg["assertion"] = assertionMsgList
				}
				if api.Regex != nil {
					debugMsg["regex"] = regex
				}
				if requestCollection != nil {
					log2.Logger.Debug("report_id", debugMsg["report_id"])
					debugMsg["report_id"] = event.ReportId
					model.Insert(requestCollection, debugMsg)
				}
			}

		case model.OnlyError:
			if isSucceed == false {
				debugMsg := make(map[string]interface{})
				debugMsg["uuid"] = api.Uuid.String()
				debugMsg["event_id"] = event.Id
				debugMsg["api_id"] = api.TargetId
				debugMsg["api_name"] = api.Name
				debugMsg["type"] = model.RequestType
				debugMsg["request_time"] = requestTime / 1000000
				debugMsg["request_code"] = resp.StatusCode()
				debugMsg["request_header"] = req.Header.String()
				if string(req.Body()) != "" {
					debugMsg["request_body"] = string(req.Body())
				} else {
					debugMsg["request_body"] = str
				}
				debugMsg["response_header"] = resp.Header.String()
				if string(resp.Body()) == "" {
					debugMsg["response_body"] = errMsg
				} else {
					debugMsg["response_body"] = string(resp.Body())
				}

				debugMsg["response_bytes"] = receivedBytes
				debugMsg["status"] = model.Failed
				debugMsg["next_list"] = event.NextList
				if api.Assert != nil {
					debugMsg["assertion"] = assertionMsgList
				}
				if api.Regex != nil {
					debugMsg["regex"] = regex
				}
				if requestCollection != nil {
					debugMsg["report_id"] = event.ReportId
					model.Insert(requestCollection, debugMsg)
				}
			}

		}
	}
	return isSucceed, errCode, requestTime, sendBytes, receivedBytes, errMsg, timestamp
}
