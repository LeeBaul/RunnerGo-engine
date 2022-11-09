package tools

import (
	"RunnerGo-engine/config"
	"RunnerGo-engine/log"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

type StopMsg struct {
	ReportId string   `json:"report_id"`
	Machines []string `json:"machines"`
}

func SendStopStressReport(machines []string, reportId string) {
	sm := StopMsg{
		ReportId: reportId,
	}
	sm.Machines = machines

	body, err := json.Marshal(&sm)
	if err != nil {
		log.Logger.Error(reportId, "   ,json转换失败：  ", err.Error())
	}
	res, err := http.Post(config.Conf.Management.Address, "application/json", strings.NewReader(string(body)))

	if err != nil {
		log.Logger.Error("http请求建立链接失败：", err.Error())
		return
	}
	defer res.Body.Close()

	_, err = ioutil.ReadAll(res.Body)

	if err != nil {
		log.Logger.Error(reportId, " ,发送停止任务失败，http请求失败", err.Error())
		return
	}
	if res.StatusCode == 200 {
		log.Logger.Error(reportId, "  :停止任务成功：")
	} else {
		log.Logger.Error(reportId, "  :停止任务失败：status code:  ", res.StatusCode)
	}

}
