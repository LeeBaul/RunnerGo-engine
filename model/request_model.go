// Package execution 请求数据模型package execution
package model

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"kp-runner/log"
	"strings"
)

// getHeaderValue 获取 header
func getHeaderValue(v string, headers map[string]string) {
	index := strings.Index(v, ":")
	if index < 0 {
		return
	}
	vIndex := index + 1
	if len(v) >= vIndex {
		value := strings.TrimPrefix(v[vIndex:], " ")
		if _, ok := headers[v[:index]]; ok {
			headers[v[:index]] = fmt.Sprintf("%s; %s", headers[v[:index]], value)
		} else {
			headers[v[:index]] = value
		}
	}
}

// TestResultDataMsg 测试结果数据结构
type TestResultDataMsg struct {
	ReportId      string
	ReportName    string
	PlanId        string // 任务ID
	PlanName      string //
	SceneId       string // 场景
	SceneName     string
	ApiId         string // 接口ID
	ApiName       string // 接口名称
	RequestTime   uint64 // 请求响应时间
	ErrorType     int    // 错误类型：1. 请求错误；2. 断言错误
	IsSucceed     bool   // 请求是否有错：true / false   为了计数
	SendBytes     int64  // 发送字节数
	ReceivedBytes int64  // 接收字节数
}

func (tr *TestResultDataMsg) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, tr); err != nil {
		log.Logger.Error("测试数据转字节失败", err)
		return nil, err
	}
	return buf.Bytes(), nil
}

func (tr *TestResultDataMsg) Length() int {
	by, _ := tr.Encode()
	return len(by)
}
