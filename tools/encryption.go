package tools

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
)

func Md5(str string) (data string) {
	hash := md5.Sum([]byte(str))
	data = fmt.Sprintf("%x", hash)
	return data
}

func EncodeMd5(str string) (data string) {

	hash := md5.Sum([]byte(str))
	data = fmt.Sprintf("%x", hash)
	return
}

func Sha256(str string) (data string) {
	h := sha256.New()
	data = fmt.Sprintf("%x", h.Sum([]byte(str)))
	return

}
