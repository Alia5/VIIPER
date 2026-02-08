package typescript

import (
	"strings"

	"github.com/Alia5/VIIPER/internal/codegen/common"
)

func goTypeToTS(goType string) string {
	base, _, _ := common.NormalizeGoType(goType)
	switch base {
	case "byte", "uint8", "uint16", "uint32", "uint64", "int8", "int16", "int32", "int64", "int", "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "string":
		return "string"
	case "any", "interface{}":
		return "unknown"
	default:
		return common.ToPascalCase(base)
	}
}

func parseGoMapType(typeStr string) (keyType string, valueType string, ok bool) {
	if !strings.HasPrefix(typeStr, "map[") {
		return "", "", false
	}
	closeIdx := strings.Index(typeStr, "]")
	if closeIdx < 0 {
		return "", "", false
	}
	keyType = typeStr[len("map["):closeIdx]
	valueType = typeStr[closeIdx+1:]
	if keyType == "" || valueType == "" {
		return "", "", false
	}
	return keyType, valueType, true
}

func writeFileHeaderTS() string { return common.FileHeader("//", "TypeScript") }
