package tools

import (
	"fmt"
	"regexp"
	"testing"
)

func TestFindDestStr(t *testing.T) {
	str := "{\"code\":10000,\"data\":{\"token\":\"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJtb2JpbGUiOiIxNTM3Mjg3NjA5MiIsInZlcl9jb2RlIjoiMTIzNCIsImV4cCI6MTY2MDY1MTY4OCwiaXNzIjoicHJvOTExIn0.D73rBvMuFiM030UyF5Mveayhe1ahpAHOtEMMwsmfN78\"},\"msg\":\"success\"}"
	rex := `{"code":\d[0-9],"data:{"`
	result := FindAllDestStr(str, rex)
	fmt.Println("111111111111", result)

	//buf := "abc azc a7c aac 888 a9c  tac"
	//compileRegex = regexp.MustCompile(`a[]0-9]c`)
	//result = compileRegex.FindAllStringSubmatch(buf, -1)
	//fmt.Println("111111111111", result)
}

func TestFindAllDestStr(t *testing.T) {
	compileRegex := regexp.MustCompile("{{(.*?)}}")
	result := compileRegex.FindAllStringSubmatch("{{token}}", -1)
	fmt.Println(result)
}
