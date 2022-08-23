package routers

import (
	"github.com/gin-gonic/gin"
	"kp-runner/api"
)

func InitRouter(Router *gin.RouterGroup) {
	{
		Router.POST("/run_plan", api.RunPlan)
		Router.GET("/run_api", api.RunApi)
		Router.POST("/run_scene", api.RunScene)
	}
}
