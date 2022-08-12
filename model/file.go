package model

import (
	"bufio"
	"io"
	"kp-runner/log"
	"os"
	"strings"
)

// ParameterizedFile 参数化文件
type ParameterizedFile struct {
	Path          string              `json:"path"`          // 文件地址
	VariableNames map[string][]string `json:"variableNames"` // 字段名称
}

// ReadFile 将参数化文件写入内存变量中
func (p *ParameterizedFile) ReadFile() {
	fs, err := os.Open(p.Path)
	defer fs.Close()
	if err != nil {
		log.Logger.Error("打开测试文件失败：", err)
		return
	}
	buf := bufio.NewReader(fs)
	i := 0
	for {
		line, err := buf.ReadString('\n')
		if err == io.EOF {
			break
		}
		line = strings.TrimSpace(line)
		var keys []string
		var value []string
		if i == 0 {
			keys = strings.Split(line, ",")
		} else {
			for j := 0; j < len(keys); j++ {
				p.VariableNames[keys[j]] = append(p.VariableNames[keys[j]], value[j])
			}
		}
		i++
	}
}
