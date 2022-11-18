package config

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"time"
)

var Conf Config

type Config struct {
	Http       Http       `yaml:"http"`
	Kafka      Kafka      `yaml:"kafka"`
	Redis1     Redis1     `yaml:"redis1"`
	Redis2     Redis2     `yaml:"redis2"`
	Mongo      Mongo      `yaml:"mongo"`
	Es         Es         `yaml:"es"`
	Grpc       Grpc       `yaml:"grpc"`
	Heartbeat  Heartbeat  `yaml:"heartbeat"`
	Oss        Oss        `yaml:"oss"`
	Machine    Machine    `yaml:"machine"`
	Management Management `yaml:"management"`
	Log        Log        `yaml:"log"`
}

type Log struct {
	Path string `yaml:"path"`
}
type Management struct {
	Address string `yaml:"address"`
}

type Machine struct {
	Bandwidth     int    `yaml:"bandwidth"`
	MaxGoroutines int    `yaml:"maxGoroutines"`
	ServerType    int    `yaml:"serverType"`
	NetName       string `yaml:"netName"`
	DiskName      string `yaml:"diskName"`
}

type Heartbeat struct {
	Port      int32  `yaml:"port"`
	Region    string `yaml:"region"`
	Duration  int64  `yaml:"duration"`
	Resources int64  `yaml:"resources"`
	GrpcHost  string `yaml:"grpcHost"`
}
type Http struct {
	Name                     string        `yaml:"name"`
	Address                  string        `yaml:"address"`
	Version                  string        `yaml:"version"`
	ReadTimeout              time.Duration `yaml:"readTimeout"`
	WriteTimeout             time.Duration `yaml:"writeTimeout"`
	MaxConnPerHost           int           `yaml:"maxConnPerHost"`
	MaxIdleConnDuration      time.Duration
	MaxConnWaitTimeout       time.Duration
	NoDefaultUserAgentHeader bool `yaml:"noDefaultUserAgentHeader"`
}

type Kafka struct {
	Address string `yaml:"address"`
	TopIc   string `yaml:"topIc"`
}

type Redis1 struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int64  `yaml:"DB"`
}
type Redis2 struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int64  `yaml:"DB"`
}

type Mongo struct {
	User             string `yaml:"user"`
	Password         string `yaml:"password"`
	Address          string `yaml:"address"`
	DB               string `yaml:"db"`
	StressDebugTable string `yaml:"stressDebugTable"`
	DebugTable       string `yaml:"debugTable"`
	SceneDebugTable  string `yaml:"sceneDebugTable"`
	ApiDebugTable    string `yaml:"apiDebugTable"`
}

type Es struct {
	UserName string `yaml:"username"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Index    string `yaml:"index"`
	Size     int    `yaml:"size"`
}

type Grpc struct {
	Port string `yaml:"port"`
}

type Oss struct {
	Endpoint        string `yaml:"endpoint"`
	Bucket          string `yaml:"bucket"`
	AccessKeyID     string `yaml:"accessKeyID"`
	AccessKeySecret string `yaml:"accessKeySecret"`
	Split           string `yaml:"split"`
	Down            string `yaml:"down"`
}

func InitConfig() {

	var conf string
	flag.StringVar(&conf, "c", "./prd.yaml", "配置文件,默认为conf文件夹下的dev文件")
	if !flag.Parsed() {
		flag.Parse()
	}

	viper.SetConfigFile(conf)
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	if err = viper.Unmarshal(&Conf); err != nil {
		panic(fmt.Errorf("unmarshal error config file: %w", err))
	}

	fmt.Println("config initialized")

}
