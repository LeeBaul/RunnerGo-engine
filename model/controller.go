package model

import (
	"strings"
	"time"
)

// Controller 控制器
type Controller struct {
	ControllerType       string                `json:"controllerType"` // wait， if， collection
	IfController         *IfController         `json:"ifController"`
	WaitController       *WaitController       `json:"waitController"`
	CollectionController *CollectionController `json:"collectionController"`
}

// IfController if控制器
type IfController struct {
	Type       string        `json:"type"`
	Name       string        `json:"name"`
	Key        string        `json:"key"`   // key，值某个变量
	Logic      string        `json:"logic"` // 逻辑运算符
	Value      string        `json:"value"` // key对应的值
	Requests   []*Request    `json:"requests"`
	Controller []*Controller `json:"controller"`
}

func (ic *IfController) PerForm(value string) {
	if ic.Requests == nil && ic.Controller == nil {
		return
	}
	switch ic.Logic {
	case Equal:
		if strings.Compare(ic.Value, value) != 0 {
			return
		}
		switch ic.Type {
		case RequestType:

		}
	case UNEqual:
		if strings.Compare(ic.Value, value) == 0 {
			return
		}

	case GreaterThan:
		if ic.Value <= value {
			return
		}

	case GreaterThanOrEqual:
		if ic.Value < value {
			return
		}

	case LessThan:
		if ic.Value >= value {
			return
		}

	case LessThanOrEqual:
		if ic.Value > value {
			return
		}
	case Includes:

	case UNIncludes:

	case NULL:

	case NotNULL:

	}

}

// WaitController 等待控制器；思考时间
type WaitController struct {
	Name       string        `json:"name"`
	WaitTime   string        `json:"waitTime"` // 等待时长，ms
	Controller []*Controller `json:"controller"`
}

// CollectionController 集合点控制器
type CollectionController struct {
	Name     string `json:"name"`
	WaitTime int64  `json:"waitTime"` // 等待多长时间，如果还没完成，则不在等待， ms
}

func (cc CollectionController) CollectionWait(list []bool, ch chan bool) {
	timeWait := cc.WaitTime
	startTime := time.Now().UnixMilli()
	for startTime+timeWait > time.Now().UnixMilli() {
		channel := <-ch
		switch channel {
		case true:
			list = append(list, true)
		case false:
			break
		}
	}
}
