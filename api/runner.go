package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"kp-runner/global"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/server"
	"net/http"
	"strconv"
)

func RunPlan(c *gin.Context) {
	var planInstance = model.Plan{}
	err := c.ShouldBindJSON(&planInstance)

	if err != nil {
		global.ReturnMsg(c, http.StatusBadRequest, "数据格式不正确", err.Error())
		return
	}

	requestJson, _ := json.Marshal(planInstance)
	log.Logger.Info("开始执行计划", string(requestJson))
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

	uid := uuid.NewV4()
	scene.Uuid = uid
	requestJson, _ := json.Marshal(scene)
	log.Logger.Info("调试场景", string(requestJson))
	go server.DebugScene(&scene)
	global.ReturnMsg(c, http.StatusOK, "调式场景", uid)
}

func RunApi(c *gin.Context) {
	var runApi = model.Api{}
	err := c.ShouldBindJSON(&runApi)
	//log.Logger.Error("body111111111111111", string(aaaa))

	if err != nil {
		global.ReturnMsg(c, http.StatusBadRequest, "数据格式不正确", err.Error())
		return
	}

	uid := uuid.NewV4()
	runApi.Uuid = uid
	runApi.Debug = model.All

	requestJson, _ := json.Marshal(&runApi)

	log.Logger.Info("调试接口", string(requestJson))
	_, _ = json.Marshal(runApi.Request.Body.Mode)
	go server.DebugApi(runApi)
	global.ReturnMsg(c, http.StatusOK, "调试接口", uid)
}

func Stop(c *gin.Context) {
	var stop model.Stop

	if err := c.ShouldBindJSON(&stop); err != nil {
		global.ReturnMsg(c, http.StatusBadRequest, "数据格式不正确", err.Error())
		return
	}
	if stop.ReportIds == nil || len(stop.ReportIds) < 1 {
		global.ReturnMsg(c, http.StatusBadRequest, "数据格式不正确", "报告列表不能为空")
		return
	}
	go func(stop model.Stop) {
		for _, reportId := range stop.ReportIds {
			err := model.InsertStatus(reportId+":status", "stop", 20)
			if err != nil {
				log.Logger.Error("向redis写入任务状态失败：", err)
			}
		}
	}(stop)
	global.ReturnMsg(c, http.StatusOK, "停止任务", nil)
}

func StopScene(c *gin.Context) {
	var stop model.StopScene
	if err := c.ShouldBindJSON(&stop); err != nil {
		global.ReturnMsg(c, http.StatusBadRequest, "数据格式不正确", err.Error())
		return
	}

	if stop.SceneId == 0 {
		global.ReturnMsg(c, http.StatusBadRequest, "scene_id不正确", stop.SceneId)
		return
	}
	stopId := strconv.FormatInt(stop.SceneId, 10)
	go func(stop model.StopScene) {
		err := model.InsertStatus(stopId+":status", "stop", 20)
		if err != nil {
			log.Logger.Error("向redis写入任务状态失败：", err)
		}
	}(stop)

	global.ReturnMsg(c, http.StatusOK, "停止成功", nil)
}
