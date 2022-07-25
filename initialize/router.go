package initialize

import (
	"github.com/gin-gonic/gin"
	"kp-runner/middlewares"
	"kp-runner/routers"
)

func Routers() *gin.Engine {

	Routers := gin.Default()

	// 配置跨域
	Routers.Use(middlewares.Cors())

	groups := Routers.Group("runner")
	routers.InitRouter(groups)

	return Routers
}
