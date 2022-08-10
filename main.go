// Package main go 实现的压测工具
package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-micro/plugins/v4/registry/consul"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/web"
	"go.uber.org/zap"
	config "kp-runner/config"
	"kp-runner/global"
	"kp-runner/initialize"
	"kp-runner/log"
	"os"
	"os/signal"
	"syscall"
)

// array 自定义数组参数
type array []string

// String string
func (a *array) String() string {
	return fmt.Sprint(*a)
}

// Set set
func (a *array) Set(s string) error {
	*a = append(*a, s)

	return nil
}

var (
	GinRouter *gin.Engine
)

func init() {

	//1. 初始化logger
	zap.S().Debug("初始化logger")
	log.InitLogger()

	//2. 初始化配置文件
	zap.S().Debug("初始化配置文件")
	config.InitConfig()

	global.ConsulClient = consul.NewRegistry(registry.Addrs(config.Config["consulAddress"].(string)))
	//3. 初始化routers
	zap.S().Debug("初始化routers")
	GinRouter = initialize.Routers()

	//4. 初始化kafka
	zap.S().Info("初始化kafka")

	//4. 语言转换
	if err := initialize.InitTrans("zh"); err != nil {
		zap.S().Debug(err)
	}

}

func main() {

	//5. 注册服务
	zap.S().Debug("注册服务")
	controllerService := web.NewService(
		web.Name(config.Config["serverName"].(string)),
		web.Version(config.Config["serverVersion"].(string)),
		web.Registry(global.ConsulClient),
		web.Address(config.Config["serverAddress"].(string)),
		web.Handler(GinRouter),
	)

	if err := controllerService.Run(); err != nil {
		return
	}

	/// 接收终止信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.S().Info("注销成功")

}
