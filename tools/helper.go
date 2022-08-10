// Package tools  工具类

package tools

import (
	"encoding/json"
	"time"
)

// TimeDifference 时间差，纳秒

func TimeDifference(startTime int64) (difference uint64) {
	difference = uint64(time.Now().UnixMilli() - startTime)
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
