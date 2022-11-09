package tools

import (
	"RunnerGo-engine/log"
	"encoding/base64"
	"strings"
)

// base64解码
func Base64DeEncode(str string, dataType string) (decoded []byte) {
	if dataType != "File" {
		return
	}
	strs := strings.Split(str, ";base64,")
	if len(strs) < 2 {
		return
	}
	str = strs[1]

	if str[len(str)-1] == 61 {
		decoded, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			log.Logger.Error("base64解码错误：", err)
		}
		return decoded
	} else {
		decoded, err := base64.RawStdEncoding.DecodeString(str)
		if err != nil {
			log.Logger.Error("base64解码错误：", err)
		}
		return decoded
	}
	return
}

// base64编码
func Base64Encode(str string) (encode string) {
	msg := []byte(str)
	encode = base64.RawStdEncoding.EncodeToString(msg)
	return
}
