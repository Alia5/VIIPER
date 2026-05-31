package common

import "strconv"

func IsIntegerConst(value interface{}, goType string) bool {
	base, _, _ := NormalizeGoType(goType)
	switch base {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "byte":
		return true
	case "string", "bool", "char", "float32", "float64":
		return false
	}

	switch v := value.(type) {
	case int, int64, uint64:
		return true
	case string:
		if _, err := strconv.ParseInt(v, 0, 64); err == nil {
			return true
		}
		if _, err := strconv.ParseUint(v, 0, 64); err == nil {
			return true
		}
	}

	return false
}
