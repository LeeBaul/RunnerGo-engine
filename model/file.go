package model

import (
	"bufio"
	"fmt"
	"github.com/valyala/fasthttp"
	"io"
	"kp-runner/config"
	"kp-runner/log"
	"kp-runner/tools"
	"os"
	"strings"
	"sync"
	"time"
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

func (p *ParameterizedFile) UseFile() {
	if p.Paths == nil || len(p.Paths) == 0 {
		return
	}
	fc := &fasthttp.Client{}
	req := fasthttp.AcquireRequest()
	// set url
	req.Header.SetMethod("GET")
	req.Header.SetMethod("GET")
	resp := fasthttp.AcquireResponse()
	defer req.ConnectionClose()
	defer resp.ConnectionClose()
	p.VariableNames.VarMapList = make(map[string][]string)
	for _, path := range p.Paths {
		req.Header.SetRequestURI(path)
		if err := fc.Do(req, resp); err != nil {
			log.Logger.Error("下载参数化文件错误：", err)
			continue
		}
		strs := strings.Split(string(resp.Body()), "\n")
		index := 0
		var keys []string
		for _, str := range strs {
			str = strings.TrimSpace(str)
			if index == 0 {
				keys = strings.Split(str, ",")
				for _, data := range keys {
					data = strings.TrimSpace(data)
					if _, ok := p.VariableNames.VarMapList[data]; !ok {
						p.VariableNames.VarMapList[data] = []string{}
					}

				}

			} else {
				dataList := strings.Split(str, ",")
				for i := 0; i < len(keys); i++ {
					data := strings.TrimSpace(dataList[i])
					p.VariableNames.VarMapList[keys[i]] = append(p.VariableNames.VarMapList[keys[i]], data)
				}
			}
			index++
		}
	}
	p.VariableNames.Index = 0
}

// DownLoadFile 下载测试文件
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
		if tools.PathExists(config.Conf.Oss.Down + "/" + teamId + reportId) {
			toPath = fmt.Sprintf("%s/%s/%s/%s", config.Conf.Oss.Down, teamId, reportId, fileName)
		} else {
			toPath = fmt.Sprintf("/data/%s", fileName)
		}
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
	fmt.Println("p:             ", p.VariableNames.VarMapList["id"])
	if values, ok := p.VariableNames.VarMapList[key]; ok {
		if p.VariableNames.Index >= len(p.VariableNames.VarMapList[key]) {
			p.VariableNames.Index = 0
		}
		time.Sleep(10 * time.Second)
		value = values[p.VariableNames.Index]
		p.VariableNames.Index++
	}
	return
}
