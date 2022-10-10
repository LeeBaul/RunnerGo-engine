package tools

import (
	"encoding/base64"
	"kp-runner/log"
	"strings"
)

// base64解码
func Base64DeEncode(str string, dataType string) (decoded []byte, fileName string) {
	if dataType != "File" {
		return
	}
	strs := strings.Split(str, ";base64,")
	if len(strs) < 2 {
		return
	}
	str = strs[1]

	fileName = strings.Split(strs[0], "/")[1]
	if str[len(str)-1] == 61 {
		decoded, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			log.Logger.Error("base64解码错误：", err)
		}
		return decoded, fileName
	} else {
		decoded, err := base64.RawStdEncoding.DecodeString(str)
		if err != nil {
			log.Logger.Error("base64解码错误：", err)
		}
		return decoded, fileName
	}
	return
}

// base64编码
func Base64Encode(str string) (encode string) {
	msg := []byte(str)
	encode = base64.RawStdEncoding.EncodeToString(msg)
	return
}
