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
	var planInstance = model.Plan{}
	err := c.ShouldBindJSON(&planInstance)

	if err != nil {
		global.ReturnMsg(c, http.StatusBadRequest, "数据格式不正确", err.Error())
		return
	}

	log.Logger.Info("开始执行计划", planInstance)
	go func(planInstance *model.Plan) {
		server.DisposeTask(planInstance)
	}(&planInstance)

	global.ReturnMsg(c, http.StatusOK, "开始执行计划", nil)

}

func RunScene(c *gin.Context) {
	//var workerInstance model.Worker
	//err := c.ShouldBindJSON(&workerInstance)
	//
	//if err != nil {
	//	global.ReturnMsg(c, http.StatusBadRequest, "数据格式不正确", err.Error())
	//	return
	//}
	//
	//log.Logger.Info("运行场景", workerInstance)
	//
	//golink.DisposeScene(workerInstance.Scene.EventList)
	//
	//global.ReturnMsg(c, http.StatusOK, "开始执行计划", nil)
}

func RunApi(c *gin.Context) {
	//var request = model.Request{}
	//err := c.ShouldBindJSON(&request)
	//
	//if err != nil {
	//	global.ReturnMsg(c, http.StatusBadRequest, "数据格式不正确", err.Error())
	//	return
	//}
	//requestResults := new(model.ResultDataMsg)
	//debugMsg := new(model.DebugMsg)
	//server.ExecutionDebugRequest(request, nil, requestResults, debugMsg)
	//global.ReturnMsg(c, http.StatusOK, "调试接口", debugMsg)
}
