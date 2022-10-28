package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"kp-runner/config"
	"kp-runner/initialize"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/server/heartbeat"
)

var (
	GinRouter *gin.Engine
)

func initService() {

	// 初始化logger
	zap.S().Debug("初始化logger")
	log.InitLogger()

	// 初始化配置文件
	zap.S().Debug("初始化配置文件")
	config.InitConfig()

	// 获取本机地址
	heartbeat.InitLocalIp()
	// 初始化redis客户端
	log.Logger.Debug("初始化redis客户端")
	if err := model.InitRedisClient(
		config.Conf.Redis.Address,
		config.Conf.Redis.Password,
		config.Conf.Redis.DB,
	); err != nil {
		log.Logger.Error("redis连接失败:", err)
		panic("redis 连接失败")
		return
	}

	//3. 初始化routers
	log.Logger.Debug("初始化routers")
	GinRouter = initialize.Routers()

	// 语言转换
	if err := initialize.InitTrans("zh"); err != nil {
		log.Logger.Error(err)
	}

	// 注册服务
	log.Logger.Debug("注册服务")
	kpRunnerService := &http.Server{
		Addr:           config.Conf.Http.Address,
		Handler:        GinRouter,
		ReadTimeout:    config.Conf.Http.ReadTimeout * time.Millisecond,
		WriteTimeout:   config.Conf.Http.WriteTimeout * time.Millisecond,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := kpRunnerService.ListenAndServe(); err != nil {
			log.Logger.Error("kpRunnerService:", err)
			return
		}
	}()
	// 注册并发送心跳数据
	field := heartbeat.LocalIp + "_" + fmt.Sprintf("%d", config.Conf.Heartbeat.Port) + "_" + config.Conf.Heartbeat.Region
	go func() {
		//heartbeat.SendHeartBeat(config.Conf.Heartbeat.GrpcHost, config.Conf.Heartbeat.Duration)
		heartbeat.SendHeartBeatRedis(field, config.Conf.Heartbeat.Duration)
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
