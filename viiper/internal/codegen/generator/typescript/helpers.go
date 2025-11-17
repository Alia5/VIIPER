package typescript

import (
	"strings"
	"viiper/internal/codegen/common"
)

// goTypeToTS maps Go types to TypeScript types
func goTypeToTS(goType string) string {
	goType = strings.TrimPrefix(goType, "*")
	goType = strings.TrimPrefix(goType, "[]")
	switch goType {
	case "uint8", "uint16", "uint32", "uint64", "int8", "int16", "int32", "int64", "int", "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "string":
		return "string"
	default:
		return common.ToPascalCase(goType)
	}
}

func writeFileHeaderTS() string {
	return "// Auto-generated VIIPER TypeScript SDK\n// DO NOT EDIT - This file is generated from the VIIPER server codebase\n\n"
}
