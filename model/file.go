package model

import (
	"bufio"
	"fmt"
	"io"
	"kp-runner/config"
	"kp-runner/log"
	"kp-runner/tools"
	"os"
	"strings"
	"sync"
)

// ParameterizedFile 参数化文件
type ParameterizedFile struct {
	Paths         []string       `json:"path"` // 文件地址
	RealPaths     []string       `json:"real_paths"`
	VariableNames *VariableNames `json:"variable_names"` // 存储变量及数据的map
}

type VariableNames struct {
	VarMapList map[string][]string `json:"var_map_list"`
	Index      int                 `json:"index"`
	Mu         sync.Mutex          `json:"mu"`
}

// 从oss下载文件
func (p *ParameterizedFile) DownLoadFile(teamId, reportId string) {
	if p.Paths == nil {
		return
	}
	client := NewOssClient(config.Conf.Oss.Endpoint, config.Conf.Oss.AccessKeyID, config.Conf.Oss.AccessKeySecret)
	if client == nil {
		return
	}
	if p.RealPaths == nil {
		p.RealPaths = []string{}
	}
	for _, path := range p.Paths {
		names := strings.Split(path, config.Conf.Oss.Split)
		if names == nil || len(names) < 2 {
			break
		}
		name := config.Conf.Oss.Split + names[1]
		files := strings.Split(name, "/")
		fileName := ""
		if len(files) > 0 {
			fileName = files[len(files)-1]
		}
		toPath := ""
		if tools.PathExists(config.Conf.Oss.Down + teamId + "/" + reportId) {
			toPath = fmt.Sprintf("%s/%s/%s/%s", config.Conf.Oss.Down, teamId, reportId, fileName)
		} else {
			toPath = fmt.Sprintf("/data/%s", fileName)
		}
		log.Logger.Debug("name................", name)
		log.Logger.Debug("topath................", toPath)
		log.Logger.Debug("config.Conf.Oss.Bucket................", config.Conf.Oss.Bucket)
		err := DownLoad(client, name, toPath, config.Conf.Oss.Bucket)
		if err != nil {
			break
		}
		p.RealPaths = append(p.RealPaths, toPath)
	}
	p.ReadFile()
}

// ReadFile 将参数化文件写入内存变量中
func (p *ParameterizedFile) ReadFile() {
	if p.RealPaths == nil {
		return
	}

	for _, file := range p.RealPaths {
		fs, err := os.Open(file)
		if err != nil {
			log.Logger.Error(file, "文件打开失败：", err)
			break
		}
		buf := bufio.NewReader(fs)
		i := 0
		p.VariableNames.VarMapList = make(map[string][]string)

		var keys []string
		for {
			line, readErr := buf.ReadString('\n')
			if readErr == io.EOF {
				break
			}

			line = strings.TrimSpace(line)

			if i == 0 {
				keys = strings.Split(line, ",")
				for _, v := range keys {
					if _, ok := p.VariableNames.VarMapList[v]; !ok {
						p.VariableNames.VarMapList[v] = []string{}
					}
				}
			} else {
				var value = strings.Split(line, ",")
				for j := 0; j < len(keys); j++ {
					p.VariableNames.VarMapList[keys[j]] = append(p.VariableNames.VarMapList[keys[j]], value[j])
				}
			}
			i++
		}
		fs.Close()
		if err = os.Remove(file); err != nil {
			log.Logger.Error("测试文件: " + file + " , 删除失败")
			break
		}
	}
	log.Logger.Debug("参数化：", p.VariableNames.VarMapList)
	p.VariableNames.Index = 0

}

func (p *ParameterizedFile) GetPathList(reportId string) {
	if p.Paths == nil && len(p.Paths) <= 0 {
		return
	}
	//for _, pathName := range p.Paths {
	//
	//}
}

// UseVar 使用数据
func (p *ParameterizedFile) UseVar(key string) (value string) {
	if values, ok := p.VariableNames.VarMapList[key]; ok {
		if p.VariableNames.Index >= len(p.VariableNames.VarMapList[key]) {
			p.VariableNames.Index = 0
		}
		value = values[p.VariableNames.Index]
		p.VariableNames.Index++
	}
	return
}
