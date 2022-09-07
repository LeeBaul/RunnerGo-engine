package config

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var Config map[string]interface{}

func InitConfig(conf string) {
	viper.SetConfigName(conf)
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

	// 读取mysql信息
	Config["mysql"] = viper.Get("mysql")

	// 读取http客户端配置
	Config["httpClientName"] = viper.Get("httpClient.name")
	Config["httpNoDefaultUserAgentHeader"] = viper.Get("httpClient.noDefaultUserAgentHeader")
	Config["httpClientMaxConnsPerHost"] = viper.Get("httpClient.maxConnsPerHost")
	Config["httpClientMaxIdleConnDuration"] = viper.Get("httpClient.maxIdleConnDuration")
	Config["httpClientReadTimeout"] = viper.Get("httpClient.writeTimeout")
	Config["httpClientWriteTimeout"] = viper.Get("httpClient.writeTimeout")
	Config["httpClientMaxConnWaitTimeout"] = viper.Get("httpClient.maxConnWaitTimeout")

	// kafka配置
	Config["kafkaAddress"] = viper.Get("kafka.address")
	Config["Topic"] = viper.Get("kafka.topic")

	// es
	Config["esHost"] = viper.Get("es.host")

	// redis
	Config["redisAddr"] = viper.Get("redis.Addr")
	Config["redisPassword"] = viper.Get("redis.password")
	Config["redisDB"] = viper.Get("redis.DB")
	Config["redisSize"] = viper.Get("redis.size")

	// grpc
	Config["GrpcPort"] = viper.Get("grpc.port")

	// mongodb
	Config["mongoUser"] = viper.Get("mongo.user")
	Config["mongoPassword"] = viper.Get("mongo.password")
	Config["mongoHost"] = viper.Get("mongo.host")
	Config["mongoDB"] = viper.Get("mongo.DB")
	Config["stressDebugTable"] = viper.Get("mongo.stressDebugTable")
	Config["sceneDebugTable"] = viper.Get("mongo.sceneDebugTable")
	Config["apiDebugTable"] = viper.Get("mongo.apiDebugTable")
	zap.S().Info(Config)
}
