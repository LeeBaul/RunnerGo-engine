package model

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"kp-runner/tools"
	"strconv"
	"strings"
)

// Assertion 断言
type Assertion struct {
	Type             int8              `json:"type"` //  0:Text; 1:Regular; 2:Json; 3:XPath
	AssertionText    *AssertionText    `json:"assertionText"`
	AssertionRegular *AssertionRegular `json:"assertionRegular"`
	AssertionJson    *AssertionJson    `json:"assertionJson"`
	AssertionXPath   *AssertionXPath   `json:"assertionXPath"`
}

// AssertionText 文本断言 0
type AssertionText struct {
	IsChecked    int    `json:"is_checked"`    // 1 选中  -1 未选
	ResponseType int8   `json:"response_type"` //  1:ResponseHeaders; 2:ResponseData; 3: ResponseCode;
	Compare      string `json:"compare"`       // Includes、UNIncludes、Equal、UNEqual、GreaterThan、GreaterThanOrEqual、LessThan、LessThanOrEqual、Includes、UNIncludes、NULL、NotNULL、OriginatingFrom、EndIn
	Var          string `json:"var"`
	Val          string `json:"val"`
}

// AssertionRegular 正则断言 1
type AssertionRegular struct {
	AssertionTarget int8   `json:"type"`       // 2:ResponseData
	Expression      string `json:"expression"` // 正则表达式

}

// AssertionJson json断言 2
type AssertionJson struct {
	Expression string `json:"expression"` // json表达式
	Condition  string `json:"condition"`  // Contain、NotContain、Equal、NotEqual
}

// AssertionXPath xpath断言 3
type AssertionXPath struct {
}

// VerifyAssertionText 验证断言 文本断言
func (assertionText *AssertionText) VerifyAssertionText(response *fasthttp.Response) (code int64, ok bool, msg string) {
	switch assertionText.ResponseType {
	case ResponseCode:
		value, err := strconv.Atoi(assertionText.Val)
		if err != nil {
			return AssertError, false, assertionText.Val + "不是int类型,转换失败"
		}
		switch assertionText.Compare {
		case Equal:
			if value == response.StatusCode() {
				return NoError, true, strconv.Itoa(response.StatusCode()) + "=" + strconv.Itoa(value) + "断言：成功"
			} else {
				return AssertError, false, strconv.Itoa(response.StatusCode()) + "不等于" + strconv.Itoa(value) + "断言：失败"
			}
		case UNEqual:
			if value != response.StatusCode() {
				return NoError, true, strconv.Itoa(response.StatusCode()) + UNEqual + strconv.Itoa(value) + "断言:成功"
			} else {
				return AssertError, false, strconv.Itoa(response.StatusCode()) + Equal + strconv.Itoa(value) + "断言:失败"
			}
		}
	case ResponseHeaders:
		switch assertionText.Compare {
		case Includes:
			if strings.Contains(response.String(), assertionText.Val) {
				return NoError, true, "响应头中包含：" + assertionText.Val + " 断言: 成功"
			} else {
				return AssertError, false, "响应头中不包含：" + assertionText.Val + " 断言: 失败"
			}
		case NULL:
			if response.Header.String() == "" {
				return NoError, true, "响应头为空， 断言: 成功"
			} else {
				return AssertError, false, "响应头不为空， 断言: 失败"
			}
		case UNIncludes:
			if strings.Contains(response.String(), assertionText.Val) {
				return AssertError, false, "响应头中包含：" + assertionText.Val + " 断言: 失败"
			} else {
				return NoError, true, "响应头中不包含：" + assertionText.Val + " 断言: 成功"

			}
		case NotNULL:
			if response.Header.String() == "" {
				return AssertError, false, "响应头不为空， 断言: 成功"

			} else {
				return NoError, true, "响应头为空， 断言: 失败"
			}
		}
	case ResponseData:
		switch assertionText.Compare {
		case Equal:
			value := tools.JsonPath(string(response.Body()), assertionText.Var)
			if value == nil {
				return AssertError, false, "响应中" + assertionText.Var + "不存在， 断言: 失败"
			}
			if num, err := strconv.ParseFloat(assertionText.Val, 64); err == nil {
				if value == num {
					return NoError, true, fmt.Sprintf("%f 等于 %f, 断言： 成功", value, num)
				} else {
					return AssertError, false, fmt.Sprintf("%f 不等于 %f, 断言： 失败", value, num)
				}
			} else {
				return AssertError, false, "不是数字类型，无法比较大小"
			}
		case UNEqual:
			value := tools.JsonPath(string(response.Body()), assertionText.Var)
			if value == nil {
				return AssertError, false, "响应中" + assertionText.Var + "不存在， 断言: 失败"
			}
			if num, err := strconv.ParseFloat(assertionText.Val, 64); err == nil {
				if value == num {
					return NoError, true, fmt.Sprintf("%f 不等于 %f, 断言： 成功", value, num)

				} else {
					return AssertError, false, fmt.Sprintf("%f 等于 %f, 断言： 失败", value, num)
				}
			} else {
				return AssertError, false, "不是数字类型，无法比较大小"
			}
		case Includes:
			if strings.Contains(response.String(), assertionText.Val) {
				return NoError, true, "响应中包含：" + assertionText.Val + " 断言: 成功"
			} else {
				return AssertError, false, "响应中不包含：" + assertionText.Val + " 断言: 失败"
			}
		case UNIncludes:
			if strings.Contains(response.String(), assertionText.Val) {
				return AssertError, false, "响应中包含：" + assertionText.Val + " 断言: 失败"
			} else {
				return NoError, true, "响应中不包含：" + assertionText.Val + " 断言: 成功"

			}
		case NULL:
			if string(response.Body()) == "" {
				return NoError, true, "响应体为空， 断言: 成功"

			} else {
				return AssertError, false, "响应不为空， 断言: 失败"
			}
		case NotNULL:
			if string(response.Body()) == "" {
				return NoError, true, "响应体不为空， 断言: 成功"
			} else {
				return AssertError, false, "响应为空， 断言: 失败"
			}

		case GreaterThan:
			value := tools.JsonPath(string(response.Body()), assertionText.Var)
			if value == nil {
				return AssertError, false, "响应中" + assertionText.Var + "不存在， 断言: 失败"
			}
			if num, err := strconv.ParseFloat(assertionText.Val, 64); err == nil {
				if value.(float64) > num {
					return NoError, true, fmt.Sprintf("%f 大于 %f, 断言： 成功", value, num)
				} else {
					return AssertError, false, fmt.Sprintf("%f 不大于 %f, 断言： 失败", value, num)
				}
			} else {
				return AssertError, false, "不是数字类型，无法比较大小"
			}
		case GreaterThanOrEqual:
			value := tools.JsonPath(string(response.Body()), assertionText.Var)
			if value == nil {
				return AssertError, false, "响应中" + assertionText.Var + "不存在， 断言: 失败"
			}
			if num, err := strconv.ParseFloat(assertionText.Val, 64); err == nil {
				if value.(float64) >= num {
					return NoError, true, fmt.Sprintf("%f 大于等于 %f, 断言： 成功", value, num)
				} else {
					return AssertError, false, fmt.Sprintf("%f 小于 %f, 断言： 失败", value, num)
				}
			} else {
				return AssertError, false, "不是数字类型，无法比较大小"
			}

		case LessThan:
			value := tools.JsonPath(string(response.Body()), assertionText.Var)
			if value == nil {
				return AssertError, false, "响应中" + assertionText.Var + "不存在， 断言: 失败"
			}
			if num, err := strconv.ParseFloat(assertionText.Val, 64); err == nil {
				if value.(float64) < num {
					return NoError, true, fmt.Sprintf("%f 小于 %f, 断言： 成功", value, num)
				} else {
					return AssertError, false, fmt.Sprintf("%f 不小于 %f, 断言： 失败", value, num)
				}
			} else {
				return AssertError, false, "不是数字类型，无法比较大小"
			}
		case LessThanOrEqual:
			value := tools.JsonPath(string(response.Body()), assertionText.Var)
			if value == nil {
				return AssertError, false, "响应中" + assertionText.Var + "不存在， 断言: 失败"
			}
			if num, err := strconv.ParseFloat(assertionText.Val, 64); err == nil {
				if value.(float64) <= num {
					return NoError, true, fmt.Sprintf("%f 小于等于 %f, 断言： 成功", value, num)
				} else {
					return AssertError, false, fmt.Sprintf("%f 大于 %f, 断言： 失败", value, num)
				}
			} else {
				return AssertError, false, "不是数字类型，无法比较大小"
			}
		}
	}
	return
}

type AssertionMsg struct {
	Type      string `json:"type"`
	Code      int64  `json:"code" bson:"code"`
	IsSucceed bool   `json:"isSucceed" bson:"isSucceed"`
	Msg       string `json:"msg" bson:"msg"`
}
