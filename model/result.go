package model

import (
	"bytes"
	"encoding/binary"
	"kp-runner/log"
)

/*
 测试结果
*/

// SceneTestResultDataMsg 场景的测试结果

type SceneTestResultDataMsg struct {
	MachineIp   string                           `json:"machineIp" bson:"machineIp"`
	MachineName string                           `json:"machineName" bson:"machineName"`
	ReportId    string                           `json:"reportId" bson:"reportId"`
	ReportName  string                           `json:"reportName" bson:"reportName"`
	PlanId      string                           `json:"planId" bson:"planId"`     // 任务ID
	PlanName    string                           `json:"planName" bson:"planName"` //
	SceneId     string                           `json:"sceneId" bson:"sceneId"`   // 场景
	SceneName   string                           `json:"sceneName" bson:"sceneName"`
	Results     map[string]*ApiTestResultDataMsg `json:"results" bson:"results"`
}

// ApiTestResultDataMsg 接口测试数据经过计算后的测试结果
type ApiTestResultDataMsg struct {
	ReportId                   string `json:"reportId"`
	ReportName                 string `json:"reportName"`
	PlanId                     string `json:"planId"`   // 任务ID
	PlanName                   string `json:"planName"` //
	SceneId                    string `json:"sceneId"`  // 场景
	SceneName                  string `json:"sceneName"`
	ApiId                      string `json:"apiId"`            // 接口ID
	ApiName                    string `json:"apiName"`          // 接口名称
	TotalRequestNum            uint64 `json:"requestNum"`       // 总请求数
	TotalRequestTime           uint64 `json:"totalRequestTime"` // 总响应时间
	SuccessNum                 uint64 `json:"successNum"`
	ErrorNum                   uint64 `json:"errorNum"`       // 错误数
	AvgRequestTime             uint64 `json:"avgRequestTime"` // 平均响应事件
	MaxRequestTime             uint64 `json:"maxRequestTime"`
	MinRequestTime             uint64 `json:"minRequestTime"` // 毫秒
	CustomRequestTimeLine      uint64 `json:"customRequestTimeLine"`
	CustomRequestTimeLineValue int64  `json:"customRequestTimeLineValue"`
	NinetyRequestTimeLine      uint64 `json:"ninetyRequestTimeLine"`
	NinetyFiveRequestTimeLine  uint64 `json:"ninetyFiveRequestTimeLine"`
	NinetyNineRequestTimeLine  uint64 `json:"ninetyNineRequestTimeLine"`
	SendBytes                  uint64 `json:"sendBytes"`     // 发送字节数
	ReceivedBytes              uint64 `json:"receivedBytes"` // 接收字节数
}

// ResultDataMsg 请求结果数据结构
type ResultDataMsg struct {
	MachineIp             string `json:"machineIp" bson:"machineIp"`
	MachineName           string `json:"machineName" bson:"machineName"`
	ReportId              string `json:"reportId" bson:"reportId"`
	ReportName            string
	EventId               string
	PlanId                string // 任务ID
	PlanName              string //
	SceneId               string // 场景
	SceneName             string
	ApiId                 string  // 接口ID
	ApiName               string  // 接口名称
	RequestTime           uint64  // 请求响应时间
	CustomRequestTimeLine int64   // 自定义响应时间线
	ErrorThreshold        float32 // 自定义错误率
	ErrorType             int64   // 错误类型：1. 请求错误；2. 断言错误
	IsSucceed             bool    // 请求是否有错：true / false   为了计数
	ErrorMsg              string  // 错误信息
	SendBytes             uint64  // 发送字节数
	ReceivedBytes         uint64  // 接收字节数
}

func (tr *ApiTestResultDataMsg) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, tr); err != nil {
		log.Logger.Error("测试数据转字节失败", err)
		return nil, err
	}
	return buf.Bytes(), nil
}

func (tr *ApiTestResultDataMsg) Length() int {
	by, _ := tr.Encode()
	return len(by)
}
