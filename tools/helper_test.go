package tools

import (
	"fmt"
	"testing"
)

func TestVariablesMatch(t *testing.T) {
	name := "{{name}}"
	fmt.Println(VariablesMatch(name))
}
