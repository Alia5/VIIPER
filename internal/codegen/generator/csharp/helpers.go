package csharp

import (
	"strings"

	"github.com/Alia5/VIIPER/internal/codegen/common"
)

func toPascalCase(s string) string {
	return common.ToPascalCase(s)
}

func toCamelCase(s string) string {
	return common.ToCamelCase(s)
}

func goTypeToCSharp(goType string) string {
	base, _, _ := common.NormalizeGoType(goType)
	if base == "any" || base == "interface{}" {
		return "object"
	}

	switch base {
	case "uint8":
		return "byte"
	case "uint16":
		return "ushort"
	case "uint32":
		return "uint"
	case "uint64":
		return "ulong"
	case "int8":
		return "sbyte"
	case "int16":
		return "short"
	case "int32", "int":
		return "int"
	case "int64":
		return "long"
	case "float32":
		return "float"
	case "float64":
		return "double"
	case "bool":
		return "bool"
	case "string":
		return "string"
	case "byte":
		return "byte"
	default:
		return toPascalCase(base)
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

func writeFileHeader() string { return common.FileHeader("//", "C#") }
