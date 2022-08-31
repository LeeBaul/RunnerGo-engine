package api

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"kp-runner/model/proto/pb"
	"reflect"
	"testing"
)

func TestGrpcServer_RunPlan(t *testing.T) {
	conn, err := grpc.Dial("192.168.110.231:9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("客户端连接grpc服务失败：", err)
		return
	}
	defer conn.Close()

	c := pb.NewGrpcServiceClient(conn)
	grpcPlan := &pb.GrpcPlan{
		PlanId:   "12345",
		PlanName: "234567",
	}

	response, err2 := c.RunPlan(context.Background(), grpcPlan)
	fmt.Println("调用RunPlan接口失败：", err2)
	if err2 != nil {
		fmt.Println("调用RunPlan接口失败：", err2)
	}

	fmt.Println(reflect.TypeOf(response).Name())
}
