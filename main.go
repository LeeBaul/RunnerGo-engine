package main

import (
	"context"
	"flag"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"kp-runner/config"
	"kp-runner/initialize"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/server/heartbeat"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	GinRouter *gin.Engine
	conf      string
)

func initService() {

	flag.StringVar(&conf, "c", "config\\runner-dev", "配置文件,默认为conf文件夹下的dev文件")
	flag.Parse()
	// 初始化logger
	zap.S().Debug("初始化logger")
	log.InitLogger()

	// 初始化配置文件
	zap.S().Debug("初始化配置文件")
	config.InitConfig(conf)

	// 获取本机地址
	heartbeat.InitLocalIp()
	// 初始化redis客户端
	log.Logger.Debug("初始化redis客户端")
	if err := model.InitRedisClient(
		config.Config["redisAddr"].(string),
		config.Config["redisPassword"].(string),
		config.Config["redisDB"].(int64),
	); err != nil {
		log.Logger.Error("redis连接失败:", err)
		return
	}

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
	kpRunnerService := &http.Server{
		Addr:           config.Config["serverAddress"].(string),
		Handler:        GinRouter,
		ReadTimeout:    time.Duration(config.Config["httpClientWriteTimeout"].(int64)) * time.Millisecond,
		WriteTimeout:   time.Duration(config.Config["httpClientWriteTimeout"].(int64)) * time.Millisecond,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := kpRunnerService.ListenAndServe(); err != nil {
			log.Logger.Error("kpRunnerService:", err)
			return
		}
	}()

	/// 接收终止信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Logger.Info("注销成功")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := kpRunnerService.Shutdown(ctx); err != nil {
		log.Logger.Info("注销成功")
	}
}

func main() {
	initService()
}
