package tools

import (
	"fmt"
	"testing"
)

func TestBreakUp(t *testing.T) {
	str := "abd in dba"
	s, str1 := BreakUp(str, "in")
	fmt.Println(s)
	fmt.Println(str1)
}
