package tools

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestEncodeMd5(t *testing.T) {
	data := Sha256("125")
	fmt.Println(data)
	fmt.Println(strconv.Itoa(time.Now().Hour()) + strconv.Itoa(time.Now().Minute()) + strconv.Itoa(time.Now().Second()))
}
