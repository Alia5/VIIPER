package typescript

import (
	"github.com/Alia5/VIIPER/internal/codegen/common"
)

func goTypeToTS(goType string) string {
	base, _, _ := common.NormalizeGoType(goType)
	switch base {
	case "uint8", "uint16", "uint32", "uint64", "int8", "int16", "int32", "int64", "int", "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "string":
		return "string"
	default:
		return common.ToPascalCase(base)
	}
}

func writeFileHeaderTS() string { return common.FileHeader("//", "TypeScript") }
