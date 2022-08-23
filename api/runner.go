package api

import (
	"github.com/gin-gonic/gin"
	"kp-runner/global"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/server"
	"net/http"
)

func RunPlan(c *gin.Context) {
	var planInstance model.Plan
	err := c.ShouldBindJSON(&planInstance)

	if err != nil {
		global.ReturnMsg(c, http.StatusBadRequest, "数据格式不正确", err.Error())
		return
	}

	log.Logger.Info("开始执行计划", planInstance)
	go func(planInstance *model.Plan) {
		server.DisposeTask(planInstance)
		if err != nil {
			global.ReturnMsg(c, http.StatusBadRequest, "计划执行失败", err.Error())
			return
		}
	}(&planInstance)

	global.ReturnMsg(c, http.StatusOK, "开始执行计划", nil)

}

func RunScene(c *gin.Context) {

}

func RunApi(c *gin.Context) {

}
