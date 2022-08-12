// Package golink 连接
package golink

import (
	"context"
	"kp-runner/tools"
	"sync"
	"time"

	pb "kp-runner/proto"

	"kp-runner/model"
	"kp-runner/server/client"
)

// Grpc grpc 接口请求
func Grpc(chanID uint64, ch chan<- *model.TestResultDataMsg, totalNumber uint64, wg *sync.WaitGroup,
	request *model.Request, ws *client.GrpcSocket) {
	defer func() {
		wg.Done()
	}()
	defer func() {
		_ = ws.Close()
	}()
	for i := uint64(0); i < totalNumber; i++ {
		grpcRequest(chanID, ch, i, request, ws)
	}
	return
}

// grpcRequest 请求
func grpcRequest(chanID uint64, ch chan<- *model.TestResultDataMsg, i uint64, request *model.Request,
	ws *client.GrpcSocket) {
	var (
		startTime = time.Now().UnixMilli()
		isSucceed = false
		errCode   = model.NoError
	)
	// 需要发送的数据
	conn := ws.GetConn()
	if conn == nil {
		errCode = model.RequestError
	} else {
		// TODO::请求接口示例
		c := pb.NewApiServerClient(conn)
		var (
			ctx = context.Background()
			req = &pb.Request{
				UserName: request.Body,
			}
		)
		rsp, err := c.HelloWorld(ctx, req)
		// fmt.Printf("rsp:%+v", rsp)
		if err != nil {
			errCode = model.RequestError
		} else {
			// 200 为成功
			if rsp.Code != 200 {
				errCode = model.RequestError
			} else {
				isSucceed = true
			}
		}
	}
	requestTime := tools.TimeDifference(startTime)
	requestResults := &model.TestResultDataMsg{
		RequestTime: requestTime,
		IsSucceed:   isSucceed,
		ErrorType:   errCode,
	}
	ch <- requestResults
}
