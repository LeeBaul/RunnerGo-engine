package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"kp-runner/global"
	"kp-runner/model/plan"
	"kp-runner/server"
	"net/http"
)

func Run(c *gin.Context) {
	var planInstance plan.Plan
	err := c.ShouldBindJSON(&planInstance)
	fmt.Println("planInstance", planInstance)
	if err != nil {
		global.ReturnMsg(c, http.StatusBadRequest, "数据格式不正确", err.Error())
		return
	}
	go func(plan.Plan) {
		server.Execution(planInstance)
		if err != nil {
			global.ReturnMsg(c, http.StatusBadRequest, "计划执行失败", err.Error())
			return
		}
	}(planInstance)

	global.ReturnMsg(c, http.StatusOK, "开始执行计划", nil)

}

func Stop(c *gin.Context) {

}

func Pause(c *gin.Context) {

}
