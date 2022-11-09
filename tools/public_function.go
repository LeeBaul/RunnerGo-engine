package tools

import (
	"fmt"
	idvalidator "github.com/guanguans/id-validator"
	uuid "github.com/satori/go.uuid"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

var ControllerMapsType = make(map[string]interface{})

// RandomFloat0 随机生成0-1之间的小数
func RandomFloat0() float64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Float64()
}

// RandomString 从list中随机生成n个字符组成的字符出
func RandomString(list []string, n int) (str string) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < n; i++ {
		index := rand.Intn(len(list) - 0)
		str = str + list[index]
	}
	return
}

// RandomInt 生成min-max之间的随机数
func RandomInt(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

// SpecifiedRandomIdCard 根据参数生成身份证号
// isEighteen 是否生成18位号码
// address    省市县三级地区官方全称：如`北京市`、`台湾省`、`香港特别行政区`、`深圳市`、`黄浦区`
// birthday   出生日期：如 `2000`、`198801`、`19990101`
// sex        性别：1为男性，0为女性
func SpecifiedRandomIdCard(isEighteen bool, address string, birthday string, sex int) string {
	return idvalidator.FakeRequireId(isEighteen, address, birthday, sex)
}

// RandomIdCard 随机生成身份证号
func RandomIdCard() string {
	return idvalidator.FakeId()
}

// VerifyIdCard 验证身份证号是否合法
func VerifyIdCard(str string, strict bool) bool {
	return idvalidator.IsValid(str, strict)
}

// ToStringLU 改变字符串大小写
func ToStringLU(str, option string) string {
	if str == "" {
		return str
	}

	switch option {
	case "L":
		return strings.ToLower(str)
	default:
		return strings.ToUpper(str)
	}
}

func GetUUid() string {
	uid := uuid.NewV4()
	return uid.String()
}

// ToTimeStamp 时间戳
func ToTimeStamp(option, t string) interface{} {
	times := time.Now()
	switch option {
	case "s":
		if t == "s" {
			return fmt.Sprintf("%d", times.Unix())
		} else {
			return times.Unix()
		}

	case "ms":
		if t == "s" {
			return fmt.Sprintf("%d", times.UnixMilli())
		} else {
			return times.UnixMilli()
		}
	case "ns":
		if t == "s" {
			return fmt.Sprintf("%d", times.UnixNano())
		} else {
			return times.UnixNano()
		}
	case "ws":
		if t == "s" {
			return fmt.Sprintf("%d", times.UnixMicro())
		} else {
			return times.UnixMicro()
		}
	}
	return ""

}

func InitPublicFunc() {
	ControllerMapsType["RandomFloat0"] = RandomFloat0
	ControllerMapsType["RandomString"] = RandomString
	ControllerMapsType["RandomInt"] = RandomInt
	ControllerMapsType["SpecifiedRandomIdCard"] = SpecifiedRandomIdCard
	ControllerMapsType["RandomIdCard"] = RandomIdCard
	ControllerMapsType["VerifyIdCard"] = VerifyIdCard
	ControllerMapsType["MD5"] = MD5
	ControllerMapsType["Sha256"] = Sha256
	ControllerMapsType["SHA512"] = SHA512
	ControllerMapsType["ToStringLU"] = ToStringLU
	ControllerMapsType["ToTimeStamp"] = ToTimeStamp
	ControllerMapsType["GetUUid"] = GetUUid
}

func CallPublicFunc(funcName string, params ...interface{}) []reflect.Value {
	if function, ok := ControllerMapsType[funcName]; ok {
		f := reflect.ValueOf(function)
		fmt.Println("f:       ", f)
		if len(params) != f.Type().NumIn() {
			return nil
		}
		in := make([]reflect.Value, len(params))
		for k, param := range params {
			in[k] = reflect.ValueOf(param)
		}
		return f.Call(in)

	}
	return nil
}
