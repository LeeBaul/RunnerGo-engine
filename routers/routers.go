package routers

import (
	"github.com/gin-gonic/gin"
	"kp-runner/api"
)

func InitRouter(Router *gin.RouterGroup) {
	{
		Router.POST("/run", api.Run)
		Router.GET("/stop", api.Stop)
		Router.POST("/pause", api.Pause)
	}
}
