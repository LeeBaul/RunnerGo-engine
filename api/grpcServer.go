package api

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/model/proto/pb"
	"kp-runner/server/heartbeat"
	"net"
	"sync"
)

type GrpcServer struct {
	*pb.UnimplementedGrpcServiceServer
}

//
//func (gs *GrpcServer) RunPlan(ctx context.Context, plan *pb.GrpcPlan) (response *pb.GrpcResponse, err error) {
//
//	response = &pb.GrpcResponse{
//		Code: 200,
//		Msg:  "开始处理任务",
//	}
//	go func(planInstance *pb.GrpcPlan) {
//		server.DisposeTask(planInstance)
//	}(plan)
//
//	return response, err
//}

func InitGrpcService(port string) {
	g := grpc.NewServer()
	reflection.Register(g)
	server := new(GrpcServer)

	log.Logger.Info("开始注册grpc服务")
	pb.RegisterGrpcServiceServer(g, server)
	listen, err := net.Listen("tcp", heartbeat.LocalIp+":"+port)
	if err != nil {
		log.Logger.Error("grpc服务监听失败:", err)
		return
	}
	log.Logger.Info("开始监听grpc服务")
	err = g.Serve(listen)
	if err != nil {
		log.Logger.Error("grpc服务启动失败:", err)
		return
	}
}

func GrpcPlanToPlan(grpcPlan *pb.GrpcPlan, plan *model.Plan) {
	plan.PlanId = grpcPlan.GetPlanId()
	plan.PlanName = grpcPlan.GetPlanName()
	plan.ReportId = grpcPlan.ReportId
	plan.ReportName = grpcPlan.ReportName
	for k, v := range grpcPlan.Variable {
		plan.Variable.Store(k, v)
	}

	configTask := new(model.ConfigTask)
	plan.ConfigTask = configTask
	plan.Scene = new(model.Scene)
	plan.Scene.SceneId = grpcPlan.Scene.SceneId
	plan.Scene.SceneName = grpcPlan.Scene.SceneName
	plan.Scene.Configuration = new(model.Configuration)
	plan.Scene.Configuration.ParameterizedFile = new(model.ParameterizedFile)
	plan.Scene.Configuration.ParameterizedFile.Path = grpcPlan.Scene.Configuration.ParameterizedFile.Path
	plan.Scene.Configuration.Variable = &sync.Map{}
	plan.Scene.Configuration.Mu = sync.Mutex{}

	plan.ConfigTask.TaskType = grpcPlan.ConfigTask.TaskType
	switch grpcPlan.ConfigTask.TaskType {
	case model.TimingTaskType:
		plan.ConfigTask.Task.TimingTask.Spec = grpcPlan.ConfigTask.Task.TimingTask.Spec
	case model.CICDTaskType:

	}

	plan.ConfigTask.TestModel = model.TestModel{}
	switch grpcPlan.ConfigTask.TestModel.Type {
	case model.RTModel:
		plan.ConfigTask.TestModel.RTTest = model.RTTest{}
		plan.ConfigTask.TestModel.RTTest.Length = grpcPlan.ConfigTask.TestModel.RtTest.Length
		plan.ConfigTask.TestModel.RTTest.TimeUp = grpcPlan.ConfigTask.TestModel.RtTest.TimeUp
		plan.ConfigTask.TestModel.RTTest.MaxConcurrent = grpcPlan.ConfigTask.TestModel.RtTest.MaxConcurrent
		plan.ConfigTask.TestModel.RTTest.LengthDuration = grpcPlan.ConfigTask.TestModel.RtTest.LengthDuration
		plan.ConfigTask.TestModel.RTTest.StableDuration = grpcPlan.ConfigTask.TestModel.RtTest.StableDuration
		plan.ConfigTask.TestModel.RTTest.StartConcurrent = grpcPlan.ConfigTask.TestModel.RtTest.StartConcurrent
	case model.ConcurrentModel:
		plan.ConfigTask.TestModel.ConcurrentTest = model.ConcurrentTest{}
		plan.ConfigTask.TestModel.ConcurrentTest.Type = grpcPlan.ConfigTask.TestModel.ConcurrentTest.Type
		plan.ConfigTask.TestModel.ConcurrentTest.TimeUp = grpcPlan.ConfigTask.TestModel.ConcurrentTest.TimeUp
		plan.ConfigTask.TestModel.ConcurrentTest.Rounds = grpcPlan.ConfigTask.TestModel.ConcurrentTest.Rounds
		plan.ConfigTask.TestModel.ConcurrentTest.Duration = grpcPlan.ConfigTask.TestModel.ConcurrentTest.Duration
		plan.ConfigTask.TestModel.ConcurrentTest.Concurrent = grpcPlan.ConfigTask.TestModel.ConcurrentTest.Concurrent
	case model.LadderModel:
		plan.ConfigTask.TestModel.LadderTest = model.LadderTest{}
		plan.ConfigTask.TestModel.LadderTest.TimeUp = grpcPlan.ConfigTask.TestModel.LadderTest.TimeUp
		plan.ConfigTask.TestModel.LadderTest.Length = grpcPlan.ConfigTask.TestModel.LadderTest.Length
		plan.ConfigTask.TestModel.LadderTest.MaxConcurrent = grpcPlan.ConfigTask.TestModel.LadderTest.MaxConcurrent
		plan.ConfigTask.TestModel.LadderTest.LengthDuration = grpcPlan.ConfigTask.TestModel.LadderTest.LengthDuration
		plan.ConfigTask.TestModel.LadderTest.StableDuration = grpcPlan.ConfigTask.TestModel.LadderTest.StableDuration
		plan.ConfigTask.TestModel.LadderTest.StartConcurrent = grpcPlan.ConfigTask.TestModel.LadderTest.StartConcurrent
	case model.ErrorRateModel:
		plan.ConfigTask.TestModel.ErrorRatTest = model.ErrorRatTest{}
		plan.ConfigTask.TestModel.ErrorRatTest.TimeUp = grpcPlan.ConfigTask.TestModel.ErrorRatTest.TimeUp
		plan.ConfigTask.TestModel.ErrorRatTest.Length = grpcPlan.ConfigTask.TestModel.ErrorRatTest.Length
		plan.ConfigTask.TestModel.ErrorRatTest.MaxConcurrent = grpcPlan.ConfigTask.TestModel.ErrorRatTest.MaxConcurrent
		plan.ConfigTask.TestModel.ErrorRatTest.LengthDuration = grpcPlan.ConfigTask.TestModel.ErrorRatTest.LengthDuration
		plan.ConfigTask.TestModel.ErrorRatTest.StableDuration = grpcPlan.ConfigTask.TestModel.ErrorRatTest.StableDuration
		plan.ConfigTask.TestModel.ErrorRatTest.StartConcurrent = grpcPlan.ConfigTask.TestModel.ErrorRatTest.StartConcurrent
	case model.QpsModel:
		plan.ConfigTask.TestModel.QpsTest = model.QpsTest{}
		plan.ConfigTask.TestModel.QpsTest.TimeUp = grpcPlan.ConfigTask.TestModel.QpsTest.TimeUp
		plan.ConfigTask.TestModel.QpsTest.Length = grpcPlan.ConfigTask.TestModel.QpsTest.Length
		plan.ConfigTask.TestModel.QpsTest.MaxConcurrent = grpcPlan.ConfigTask.TestModel.QpsTest.MaxConcurrent
		plan.ConfigTask.TestModel.QpsTest.LengthDuration = grpcPlan.ConfigTask.TestModel.QpsTest.LengthDuration
		plan.ConfigTask.TestModel.QpsTest.StableDuration = grpcPlan.ConfigTask.TestModel.QpsTest.StableDuration
		plan.ConfigTask.TestModel.QpsTest.StartConcurrent = grpcPlan.ConfigTask.TestModel.QpsTest.StartConcurrent
	case model.TpsModel:
		plan.ConfigTask.TestModel.TpsTest = model.TpsTest{}
		plan.ConfigTask.TestModel.TpsTest.TimeUp = grpcPlan.ConfigTask.TestModel.TpsTest.TimeUp
		plan.ConfigTask.TestModel.TpsTest.Length = grpcPlan.ConfigTask.TestModel.TpsTest.Length
		plan.ConfigTask.TestModel.TpsTest.Threshold = grpcPlan.ConfigTask.TestModel.TpsTest.Threshold
		plan.ConfigTask.TestModel.TpsTest.MaxConcurrent = grpcPlan.ConfigTask.TestModel.TpsTest.MaxConcurrent
		plan.ConfigTask.TestModel.TpsTest.StableDuration = grpcPlan.ConfigTask.TestModel.TpsTest.StableDuration
		plan.ConfigTask.TestModel.TpsTest.LengthDuration = grpcPlan.ConfigTask.TestModel.TpsTest.LengthDuration
		plan.ConfigTask.TestModel.TpsTest.StartConcurrent = grpcPlan.ConfigTask.TestModel.TpsTest.StartConcurrent
	}

	plan.Scene.EventList = []model.Event{}
	for _, value := range grpcPlan.Scene.EventList {
		switch value.EventType {
		case model.RequestType:
			request := model.Request{}
			grpcRequestToRequest(value.Request, request)
			event := model.Event{}
			event.EventType = "request"
			event.Request = request
			plan.Scene.EventList = append(plan.Scene.EventList, event)
		case model.ControllerType:

		}
	}

}

func grpcRequestToRequest(grpcRequest *pb.Request, request model.Request) {

	request.ApiId = grpcRequest.ApiId
	request.ApiName = grpcRequest.ApiName
	request.Form = grpcRequest.Form
	request.Body = grpcRequest.Body
	request.URL = grpcRequest.Url
	request.CustomRequestTime = grpcRequest.CustomRequestTime
	request.Debug = grpcRequest.Debug
	request.Weight = grpcRequest.Weight
	request.ErrorThreshold = grpcRequest.ErrorThreshold
	request.Headers = grpcRequest.Headers
	request.RequestTimeThreshold = grpcRequest.RequestTimeThreshold
	request.Parameterizes = new(sync.Map)
	request.Requests = []model.Request{}

	for k, v := range grpcRequest.Parameterizes {
		request.Parameterizes.Store(k, v)
	}
	for _, v := range grpcRequest.Requests {
		newRequest := model.Request{}
		grpcRequestToRequest(v, newRequest)
	}

}
