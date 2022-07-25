package global

import (
	ut "github.com/go-playground/universal-translator"
	"go-micro.dev/v4/registry"
)

var (
	Trans        ut.Translator
	ConsulClient registry.Registry
)
