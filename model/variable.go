package model

import "sync"

// Variable 全局
type Variable struct {
	// 全局变量
	VariableMap *sync.Map `json:"variableMap"`
}
