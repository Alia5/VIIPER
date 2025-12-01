package cpp

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/Alia5/VIIPER/internal/codegen/common"
	"github.com/Alia5/VIIPER/internal/codegen/meta"
	"github.com/Alia5/VIIPER/internal/codegen/scanner"
)

func tplFuncs(md *meta.Metadata) template.FuncMap {
	return template.FuncMap{
		"pascalcase":           common.ToPascalCase,
		"camelcase":            common.ToCamelCase,
		"snakecase":            common.ToSnakeCase,
		"toScreamingSnakeCase": func(s string) string { return strings.ToUpper(common.ToSnakeCase(s)) },
		"upper":                strings.ToUpper,
		"lower":                strings.ToLower,
		"cpptype":              cppType,
		"unwrapOptional":       func(t string) string { return strings.TrimSuffix(strings.TrimPrefix(t, "std::optional<"), ">") },
		"sliceElementType": func(t string) string {
			return strings.TrimSuffix(strings.TrimPrefix(t, "std::vector<"), ">")
		},
		"isCustomType": isCustomType,
		"hasWireTag":   func(device, dir string) bool { return common.HasWireTag(md, device, dir) },
		"wireFields":   func(device, dir string) []scanner.WireField { return common.GetWireFields(md, device, dir) },
		"isArrayType":  func(t string) bool { return strings.Contains(t, "*") },
		"baseType":     func(t string) string { return strings.Split(t, "*")[0] },
		"arrayCountField": func(t string) string {
			parts := strings.Split(t, "*")
			if len(parts) == 2 {
				return parts[1]
			}
			return ""
		},
		"isCountField": func(fields []scanner.WireField, fieldName string) bool {
			for _, f := range fields {
				if strings.Contains(f.Type, "*") {
					parts := strings.Split(f.Type, "*")
					if len(parts) == 2 && parts[1] == fieldName {
						return true
					}
				}
			}
			return false
		},
		"pathParams":           common.ExtractPathParams,
		"pathParamType":        pathParamType,
		"formatPathParamValue": formatPathParamValue,
		"payloadCppType":       func(pi scanner.PayloadInfo) string { return payloadCppType(pi) },
		"responseCppType":      func(name string) string { return common.ToPascalCase(name) },
		"isByteKeyMap":         isByteKeyMap,
		"hasCharLiteralKeys":   hasCharLiteralKeys,
		"isNumericMapVal":      func(vt string) bool { return vt != "string" && vt != "bool" },
		"sortedEntries":        common.SortedMapEntries,
		"formatKey":            formatKey,
		"formatValue":          formatValue,
	}
}

func cppType(goType string) string {
	base, isSlice, isPointer := common.NormalizeGoType(goType)

	cppBase := goBaseToCpp(base)
	if isSlice {
		return "std::vector<" + cppBase + ">"
	}
	if isPointer {
		return "std::optional<" + cppBase + ">"
	}
	return cppBase
}

func goBaseToCpp(base string) string {
	switch base {
	case "u8", "uint8", "byte":
		return "std::uint8_t"
	case "i8", "int8":
		return "std::int8_t"
	case "u16", "uint16":
		return "std::uint16_t"
	case "i16", "int16":
		return "std::int16_t"
	case "u32", "uint32":
		return "std::uint32_t"
	case "i32", "int32", "int":
		return "std::int32_t"
	case "u64", "uint64":
		return "std::uint64_t"
	case "i64", "int64":
		return "std::int64_t"
	case "float32":
		return "float"
	case "float64":
		return "double"
	case "bool":
		return "bool"
	case "string":
		return "std::string"
	default:
		return common.ToPascalCase(base)
	}
}

func writeFileHeader() string {
	return common.FileHeader("//", "C++")
}

func pathParamType(paramName string) string {
	lower := strings.ToLower(paramName)
	if strings.Contains(lower, "deviceid") || strings.Contains(lower, "devid") {
		return "const std::string&"
	}
	return "std::uint32_t"
}

func formatPathParamValue(paramName string) string {
	lower := strings.ToLower(paramName)
	if strings.Contains(lower, "deviceid") || strings.Contains(lower, "devid") {
		return paramName
	}
	return "std::to_string(" + paramName + ")"
}

func payloadCppType(pi scanner.PayloadInfo) string {
	switch pi.Kind {
	case scanner.PayloadJSON:
		if pi.RawType != "" {
			return "const " + common.ToPascalCase(pi.RawType) + "&"
		}
		return "const std::string&"
	case scanner.PayloadNumeric:
		if pi.Required {
			return "std::uint32_t"
		}
		return "std::optional<std::uint32_t>"
	case scanner.PayloadString:
		return "const std::string&"
	default:
		return ""
	}
}

func isByteKeyMap(keyType string) bool {
	switch keyType {
	case "byte", "uint8", "int8", "rune", "char":
		return true
	default:
		return false
	}
}

func hasCharLiteralKeys(entries map[string]interface{}) bool {
	for key := range entries {
		if len(key) == 1 || (len(key) == 2 && key[0] == '\\') {
			return true
		}
	}
	return false
}

func isCustomType(goType string) bool {
	base, _, _ := common.NormalizeGoType(goType)
	switch base {
	case "uint8", "byte", "uint16", "uint32", "uint64",
		"int8", "int16", "int32", "int64", "int",
		"float32", "float64", "bool", "string",
		"u8", "u16", "u32", "u64", "i8", "i16", "i32", "i64":
		return false
	default:
		return true
	}
}

func formatKey(key string, isByteKey bool) string {
	if isByteKey && (len(key) == 1 || (len(key) == 2 && key[0] == '\\')) {
		if len(key) == 2 && key[0] == '\\' {
			switch key[1] {
			case 'n':
				return "static_cast<std::uint8_t>(0x0A)"
			case 'r':
				return "static_cast<std::uint8_t>(0x0D)"
			case 't':
				return "static_cast<std::uint8_t>(0x09)"
			case '\\':
				return "static_cast<std::uint8_t>(0x5C)"
			case '\'':
				return "static_cast<std::uint8_t>(0x27)"
			}
		}
		return fmt.Sprintf("static_cast<std::uint8_t>(0x%02X)", key[0])
	}
	return key
}

func formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		if len(v) > 0 && v[0] >= 'A' && v[0] <= 'Z' {
			return v
		}
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}
