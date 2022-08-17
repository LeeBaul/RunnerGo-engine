package model

import (
	"bufio"
	"io"
	"kp-runner/log"
	"os"
	"strings"
	"sync"
)

// ParameterizedFile 参数化文件
type ParameterizedFile struct {
	Path          string         `json:"path"`          // 文件地址
	VariableNames *VariableNames `json:"variableNames"` // 存储变量及数据的map
}

type VariableNames struct {
	VarMapList map[string][]string `json:"varMapList"`
	Index      int                 `json:"index"`
	Mu         sync.Mutex          `json:"mu"`
}

// ReadFile 将参数化文件写入内存变量中
func (p *ParameterizedFile) ReadFile() {
	fs, err := os.Open(p.Path)
	defer fs.Close()
	if err != nil {
		log.Logger.Error("打开参数化文件失败：", err)
		return
	}
	buf := bufio.NewReader(fs)
	i := 0
	p.VariableNames.VarMapList = make(map[string][]string)

	var keys []string
	for {
		line, err := buf.ReadString('\n')
		if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)

		if i == 0 {
			keys = strings.Split(line, ",")
			for _, v := range keys {
				p.VariableNames.VarMapList[v] = []string{}
			}
		} else {
			var value = strings.Split(line, ",")

			for j := 0; j < len(keys); j++ {
				p.VariableNames.VarMapList[keys[j]] = append(p.VariableNames.VarMapList[keys[j]], value[j])
			}
		}
		i++
	}
	p.VariableNames.Index = 0
}

// UseVar 使用数据
func (p *ParameterizedFile) UseVar(key string) (value string) {
	if values, ok := p.VariableNames.VarMapList[key]; ok {
		if p.VariableNames.Index >= len(p.VariableNames.VarMapList[key]) {
			//p.VariableNames.Mu.Lock()
			p.VariableNames.Index = 0
			//p.VariableNames.Mu.Unlock()
		}
		value = values[p.VariableNames.Index]
		//p.VariableNames.Mu.Lock()
		p.VariableNames.Index++
		//p.VariableNames.Mu.Unlock()
	}
	return
}
