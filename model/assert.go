package model

import (
	"github.com/valyala/fasthttp"
	"kp-runner/log"
	"strconv"
)

// Assertion 断言
type Assertion struct {
	Type             int8             `json:"type"` //  0:Text; 1:Regular; 2:Json; 3:XPath
	AssertionText    AssertionText    `json:"assertionText"`
	AssertionRegular AssertionRegular `json:"assertionRegular"`
	AssertionJson    AssertionJson    `json:"assertionJson"`
	AssertionXPath   AssertionXPath   `json:"assertionXPath"`
}

// AssertionText 文本断言 0
type AssertionText struct {
	AssertionTarget int8   `json:"AssertionTarget"` // 0: ResponseCode; 1:ResponseHeaders; 2:ResponseData
	Condition       string `json:"condition"`       // Includes、UNIncludes、Equal、UNEqual、GreaterThan、GreaterThanOrEqual、LessThan、LessThanOrEqual、Includes、UNIncludes、NULL、NotNULL、OriginatingFrom、EndIn
	Value           string `json:"value"`
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

// VerifyAssertionText 验证断言
func (assertionText *AssertionText) VerifyAssertionText(response *fasthttp.Response) (code int, ok bool) {
	switch assertionText.AssertionTarget {
	case ResponseCode:
		value, err := strconv.Atoi(assertionText.Value)
		if err != nil {
			log.Logger.Info(assertionText.Value, "不是int类型,转换失败")
			return AssertError, false
		}
		switch assertionText.Condition {
		case Equal:
			if value == response.StatusCode() {
				log.Logger.Info(strconv.Itoa(response.StatusCode()) + "=" + strconv.Itoa(value) + "断言：成功")
				return NoError, true
			} else {
				log.Logger.Info(strconv.Itoa(response.StatusCode()) + "不等于" + strconv.Itoa(value) + "断言：失败")
				return AssertError, false
			}
		case UNEqual:
			if value == response.StatusCode() {
				log.Logger.Info(strconv.Itoa(response.StatusCode()) + UNEqual + strconv.Itoa(value) + "断言:成功")
				return NoError, true
			} else {
				log.Logger.Info(strconv.Itoa(response.StatusCode()) + Equal + strconv.Itoa(value) + "断言:失败")
				return AssertError, false
			}
		}
	case ResponseHeaders:
		switch assertionText.Condition {
		case Includes:

		}
	case ResponseData:

	}
	return
}
