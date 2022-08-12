package tools

import (
	"fmt"
	"testing"
)

func get(map1 map[string]string) {
	map1["name"] = "1"
}
func TestBreakUp(t *testing.T) {
	map1 := make(map[string]string)
	get(map1)

	fmt.Println("1111111111", map1)
	str := "abd in dba"
	s, str1 := BreakUp(str, "in")
	fmt.Println(s)
	fmt.Println(str1)
}
