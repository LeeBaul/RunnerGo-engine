package plugins

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	example "github.com/hashicorp/go-plugin/examples/basic/commons"
	"github.com/hashicorp/go-plugin/examples/grpc/shared"
	"kp-runner/log"
	"os"
	"os/exec"
)

// 插件库

var PluginClient plugin.Client

func NewPluginClient(pluginName string) {

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugins",
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	PluginClient := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: shared.Handshake,
		Plugins:         shared.PluginMap,
		Cmd:             exec.Command("sh", "-c", os.Getenv("KV_PLUGIN")),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC,
			plugin.ProtocolGRPC,
		},
		Logger: logger,
	})

	rpcClient, err := PluginClient.Client()
	if err != nil {
		log.Logger.Info("协议客户端创建失败：", err)
	}

	raw, err := rpcClient.Dispense(pluginName)
	if err != nil {
		log.Logger.Info("插件实例创建失败：", err)
	}

	// 像调用普通函数一样调用接口函数就ok，很方便是不是？
	greeter := raw.(example.Greeter)
	greeter.Greet()
}
