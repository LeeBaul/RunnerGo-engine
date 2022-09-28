// Package tools  工具类

package tools

import (
	"encoding/json"
	"kp-runner/log"
	"os"
	"regexp"
	"strings"
	"time"
)

// TimeDifference 时间差，纳秒

func TimeDifference(startTime int64) (difference uint64) {
	difference = uint64(time.Now().UnixNano() - startTime)
	return
}

// InArrayStr 判断字符串是否在数组内
func InArrayStr(str string, arr []string) (inArray bool) {
	for _, s := range arr {
		if s == str {
			inArray = true
			break
		}
	}
	return
}

// ToString 将map转化为string
func ToString(args map[string]interface{}) string {
	str, _ := json.Marshal(args)
	return string(str)
}

var SymbolList = []string{"`", "?", "~", "\\", "&", "*", "^", "%", "$", "￥", "#", "@", "!", "=", "+", "-", "_", "(", ")", "<", ">", ",", "."}
var PreFix = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

// VariablesMatch 变量匹配, 如 name = {{name}}
func VariablesMatch(str string) (value string) {
	if strings.Contains(str, "{{") && strings.Contains(str, "}}") && strings.Index(str, "}}") > strings.Index(str, "{{") {
		value = FindDestStr(str, "{{(.*?)}}")
		for _, v := range PreFix {
			if strings.HasPrefix(value, v) {
				return str
			}
		}
		for _, v := range SymbolList {
			if strings.Contains(value, v) {
				return str
			}
		}
		return value
	}
	return str
}

// FindDestStr 匹配规则
func FindDestStr(str string, rex string) string {
	compileRegex := regexp.MustCompile(rex)
	matchArr := compileRegex.FindStringSubmatch(str)

	if len(matchArr) > 0 {
		return matchArr[len(matchArr)-1]
	}
	return ""
}

// FindAllDestStr 匹配所有的
func FindAllDestStr(str, rex string) (result [][]string) {
	compileRegex := regexp.MustCompile(rex)
	result = compileRegex.FindAllStringSubmatch(str, -1)
	return
}

// PathExists 判断文件或文件夹是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			mkErr := os.MkdirAll(path, os.ModePerm)
			if mkErr != nil {
				log.Logger.Error("创建文件夹失败")
				return false
			}
		}
		return false
	}
	if os.IsNotExist(err) {
		mkErr := os.MkdirAll(path, os.ModePerm)
		if mkErr != nil {
			log.Logger.Error("创建文件夹失败")
			return false
		}
	}
	return true

}
