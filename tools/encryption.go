package tools

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
)

type Encryption interface {
	HashFunc(str string) (data string)
}

type Md5Type struct {
}

func (md Md5Type) HashFunc(str string) string {
	return MD5(str)
}

type Sha256Type struct {
}

func (sh Sha256Type) HashFunc(str string) string {
	return Sha256(str)
}

type Sha512Type struct {
}

func (sh Sha512Type) HashFunc(str string) string {
	return SHA512(str)
}

func GetEncryption(str string) (encryption Encryption) {
	if str == "MD5" || str == "MD5-sess" {
		encryption = Md5Type{}
		return
	}
	if str == "SHA-256" || str == "SHA-256-sess" {
		encryption = Sha256Type{}
		return
	}
	if str == "SHA-512-256" || str == "SHA-512-256-sess" {
		encryption = Sha512Type{}
	}
	return

}

func MD5(str string) string {
	hash := md5.Sum([]byte(str))
	return fmt.Sprintf("%x", hash)
}

func Sha256(str string) string {
	h := sha256.New()
	return fmt.Sprintf("%x", h.Sum([]byte(str)))
}

func SHA512(str string) string {
	h := sha512.New()
	return fmt.Sprintf("%x", h.Sum([]byte(str)))
}
