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
	concurrency uint64 = 1       // 并发数
	totalNumber uint64 = 1       // 请求数(单个并发/协程)
	debugStr           = "false" // 是否是debug
	requestURL         = ""      // 压测的url 目前支持，http/https ws/wss
	path               = ""      // curl文件路径 http接口压测，自定义参数设置
	verify             = ""      // verify 验证方法 在server/verify中 http 支持:statusCode、json webSocket支持:json
	headers     array            // 自定义头信息传递给服务器
	body        = ""             // HTTP POST方式传送数据
	GinRouter   *gin.Engine
)

func init() {
	//flag.Uint64Var(&concurrency, "c", concurrency, "并发数")
	//flag.Uint64Var(&totalNumber, "n", totalNumber, "请求数(单个并发/协程)")
	//flag.StringVar(&debugStr, "d", debugStr, "调试模式")
	//flag.StringVar(&requestURL, "u", requestURL, "压测地址")
	//flag.StringVar(&path, "p", path, "curl文件路径")
	//flag.StringVar(&verify, "v", verify, "验证方法 http 支持:statusCode、json webSocket支持:json")
	//flag.Var(&headers, "H", "自定义头信息传递给服务器 示例:-H 'Content-Type: application/json'")
	//flag.StringVar(&body, "data", body, "HTTP POST方式传送数据")
	//// 解析参数
	//flag.Parse()

	//1. 初始化logger
	zap.S().Debug("初始化logger")
	initialize.InitLogger()

	//2. 初始化配置文件
	zap.S().Debug("初始化配置文件")
	config.InitConfig()

	global.ConsulClient = consul.NewRegistry(registry.Addrs(config.Config["consulAddress"].(string)))
	//3. 初始化routers
	zap.S().Debug("初始化routers")
	GinRouter = initialize.Routers()

	//4. 语言转换
	if err := initialize.InitTrans("zh"); err != nil {
		zap.S().Debug(err)
	}

}

// main go 实现的压测工具
// 编译可执行文件
//go:generate go build main.go
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

	//if concurrency == 0 || totalNumber == 0 || (requestURL == "" && path == "") {
	//	fmt.Printf("示例: go run main.go -c 1 -n 1 -u https://www.baidu.com/ \n")
	//	fmt.Printf("压测地址或curl路径必填 \n")
	//	fmt.Printf("当前请求参数: -c %d -n %d -d %v -u %s \n", concurrency, totalNumber, debugStr, requestURL)
	//	flag.Usage()
	//	return
	//}
	//debug := strings.ToLower(debugStr) == "true"
	//request, err := model.NewRequest(requestURL, verify, 0, debug, path, headers, body)
	//if err != nil {
	//	fmt.Printf("参数不合法 %v \n", err)
	//	return
	//}
	//fmt.Printf("\n 开始启动  并发数:%d 请求数:%d 请求参数: \n", concurrency, totalNumber)
	//request.Print()
	//// 开始处理
	//server.Dispose(concurrency, totalNumber, request)
	//return
}
