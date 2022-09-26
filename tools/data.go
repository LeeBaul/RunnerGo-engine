package tools

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// base64解码
func Base64DeEncode(str string, dataType string) (decoded []byte) {
	if dataType == "file" {
		str = strings.Split(str, ";base64,")[1]
	}

	decoded, err := base64.RawStdEncoding.DecodeString(str)
	if err != nil {
		fmt.Println("base64解码错误：", err)
		return nil
	}
	return
}

// base64编码
func Base64Encode(str string) (encode string) {
	msg := []byte(str)
	encode = base64.RawStdEncoding.EncodeToString(msg)
	return
}
