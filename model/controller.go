package model

import (
	"strings"
)

func (ic *Event) PerForm(value string) {
	if ic.PreList == nil || len(ic.PreList) <= 0 {
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
