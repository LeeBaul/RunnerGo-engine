package routers

import (
	"RunnerGo-engine/api"
	"github.com/gin-gonic/gin"
)

func InitRouter(Router *gin.RouterGroup) {
	{
		Router.POST("/run_plan/", api.RunPlan)
		Router.POST("/run_api/", api.RunApi)
		Router.POST("/run_scene/", api.RunScene)
		//Router.POST("/stop/", api.Stop)
		//Router.POST("/stop_scene/", api.StopScene)
	}
}
