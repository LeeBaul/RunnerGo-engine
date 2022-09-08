package api

import (
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
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
	var scene model.Scene
	err := c.ShouldBindJSON(&scene)
	if err != nil {
		global.ReturnMsg(c, http.StatusBadRequest, "数据格式不正确", err.Error())
		return
	}

	log.Logger.Info("运行场景", scene)
	uid := uuid.NewV4()
	scene.Uuid = uid
	go server.DebugScene(&scene)

	global.ReturnMsg(c, http.StatusOK, "调式场景", uid)
}

func RunApi(c *gin.Context) {
	var api = model.Api{}
	err := c.ShouldBindJSON(&api)
	if err != nil {
		global.ReturnMsg(c, http.StatusBadRequest, "数据格式不正确", err.Error())
		return
	}
	uid := uuid.NewV4()
	api.Uuid = uid
	api.Debug = true
	go server.DebugApi(api)
	global.ReturnMsg(c, http.StatusOK, "调试接口", uid)
}

func GetId(c *gin.Context) {

}
