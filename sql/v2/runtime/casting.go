package runtime

import (
	"fmt"
)

func Cast(val interface{}, target Type) (interface{}, error) {
	if target.IsSameType(val) {
		return val, nil
	}

	return nil, fmt.Errorf("undefined cast from %v to %v", TypeFromVal(val), target)
}

func CanCast(val interface{}, target Type) bool {
	return false
}
