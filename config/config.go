package config

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var Config map[string]interface{}

func InitConfig() {
	viper.SetConfigName("config\\runner")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		zap.S().Error("配置文件打开失败：", err)
		return
	}

	Config = make(map[string]interface{})
	// 读取服务相关配置信息
	Config["serverName"] = viper.Get("server.name")
	Config["serverAddress"] = viper.Get("server.address")
	Config["serverVersion"] = viper.Get("server.version")
	Config["consulAddress"] = viper.Get("consul.address")

	// 读取mysql信息
	Config["mysql"] = viper.Get("mysql")

	// 读取http客户端配置
	Config["httpClientName"] = viper.Get("httpClient.name")
	Config["httpNoDefaultUserAgentHeader"] = viper.Get("httpClient.noDefaultUserAgentHeader")
	Config["httpClientMaxConnsPerHost"] = viper.Get("httpClient.maxConnsPerHost")
	Config["httpClientMaxIdleConnDuration"] = viper.Get("httpClient.maxIdleConnDuration")
	Config["httpClientWriteTimeout"] = viper.Get("httpClient.writeTimeout")
	Config["httpClientMaxConnWaitTimeout"] = viper.Get("httpClient.maxConnWaitTimeout")
	zap.S().Info(Config)
}
