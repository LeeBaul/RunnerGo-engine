package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-micro/plugins/v4/registry/consul"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/web"
	"go.uber.org/zap"
	"kp-runner/config"
	"kp-runner/global"
	"kp-runner/initialize"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/server/heartbeat"
	"os"
	"os/signal"
	"syscall"
)

var (
	GinRouter *gin.Engine
)

func init() {

	// 初始化logger
	zap.S().Debug("初始化logger")
	log.InitLogger()

	// 初始化配置文件
	zap.S().Debug("初始化配置文件")
	config.InitConfig()

	// 获取本机地址
	heartbeat.InitLocalIp()
	// 初始化redis客户端
	if err := model.InitRedisClient(
		config.Config["redisAddr"].(string),
		config.Config["redisPassword"].(string),
		config.Config["redisDB"].(int64),
		config.Config["redisSize"].(int64),
	); err != nil {
		log.Logger.Error("redis连接失败:", err)
		return
	}

	global.ConsulClient = consul.NewRegistry(registry.Addrs(config.Config["consulAddress"].(string)))
	//3. 初始化routers
	log.Logger.Debug("初始化routers")
	GinRouter = initialize.Routers()

	// 语言转换
	if err := initialize.InitTrans("zh"); err != nil {
		log.Logger.Error(err)
	}

	//go func() {
	//	log.Logger.Debug("注册grpc服务")
	//	api.InitGrpcService(config.Config["GrpcPort"].(string))
	//	fmt.Println("000000000000000000000000000")
	//}()

	// 注册服务
	log.Logger.Debug("注册服务")
	kpRunnerService := web.NewService(
		web.Name(config.Config["serverName"].(string)),
		web.Version(config.Config["serverVersion"].(string)),
		web.Registry(global.ConsulClient),
		web.Address(config.Config["serverAddress"].(string)),
		web.Handler(GinRouter),
	)

	if err := kpRunnerService.Run(); err != nil {
		log.Logger.Error("kpRunnerService启动失败：", err)
		return
	}

}

func main() {

	/// 接收终止信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Logger.Info("注销成功")
}
