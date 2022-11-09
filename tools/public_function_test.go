package tools

import (
	"fmt"
	"testing"
)

func TestCallPublicFunc(t *testing.T) {
	InitPublicFunc()
	fmt.Println(CallPublicFunc("RandomFloat0")[0])
	fmt.Println("md5:      ", MD5("ABC"))
}
