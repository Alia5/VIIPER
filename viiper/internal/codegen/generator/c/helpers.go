package cgen

import (
	"fmt"
	"strings"
	"text/template"
	"viiper/internal/codegen/common"
	"viiper/internal/codegen/meta"
	"viiper/internal/codegen/scanner"
)

func tplFuncs(md *meta.Metadata) template.FuncMap {
	return template.FuncMap{
		"ctype":          cType,
		"snakecase":      common.ToSnakeCase,
		"upper":          strings.ToUpper,
		"hasWireTag":     func(device, direction string) bool { return hasWireTag(md, device, direction) },
		"wireFields":     func(device, direction string) string { return wireFields(md, device, direction) },
		"indent":         indent,
		"fieldDecl":      func(f scanner.FieldInfo) string { return fieldDecl(md, f) },
		"pathParams":     orderedPathParams,
		"join":           strings.Join,
		"mapFuncDecl":    mapFuncDecl,
		"mapFuncImpl":    mapFuncImpl,
		"payloadCType":   func(pi scanner.PayloadInfo) string { return payloadCType(md, pi) },
		"responseCType":  func(name string) string { return responseCType(md, name) },
		"marshalPayload": func(pi scanner.PayloadInfo) string { return marshalPayload(md, pi) },
		"genFreeFunc":    func(dto scanner.DTOSchema) string { return generateFreeFunction(md, dto) },
		"genParser":      func(dto scanner.DTOSchema) string { return generateParser(md, dto) },
	}
}

func payloadCType(md *meta.Metadata, pi scanner.PayloadInfo) string {
	switch pi.Kind {
	case scanner.PayloadJSON:
		if pi.RawType != "" {
			return fmt.Sprintf("const viiper_%s_t*", dtoToCTypeName(md, pi.RawType))
		}
		return "const char*" // fallback to raw JSON string
	case scanner.PayloadNumeric:
		if pi.Required {
			return "uint32_t"
		}
		return "uint32_t*"
	case scanner.PayloadString:
		return "const char*"
	default:
		return ""
	}
}

func responseCType(md *meta.Metadata, dtoName string) string {
	if dtoName == "" {
		return ""
	}
	return fmt.Sprintf("viiper_%s_t", dtoToCTypeName(md, dtoName))
}

// dtoToCTypeName converts a DTO name to its C typedef name using metadata mappings
func dtoToCTypeName(md *meta.Metadata, dtoName string) string {
	// Check if there's an explicit mapping for this DTO
	if md.CTypeNames != nil {
		if mapped, ok := md.CTypeNames[dtoName]; ok {
			return mapped
		}
	}
	// Default: just convert to snake_case
	return common.ToSnakeCase(dtoName)
}

// marshalPayload generates C code to marshal a struct to JSON string
func marshalPayload(md *meta.Metadata, pi scanner.PayloadInfo) string {
	if pi.Kind != scanner.PayloadJSON || pi.RawType == "" {
		return ""
	}

	// Find the DTO in metadata
	var dto *scanner.DTOSchema
	for i := range md.DTOs {
		if md.DTOs[i].Name == pi.RawType {
			dto = &md.DTOs[i]
			break
		}
	}
	if dto == nil {
		return ""
	}

	varName := "request"
	lines := []string{
		fmt.Sprintf("if (%s) {", varName),
	}

	// Start JSON object
	firstField := true
	for _, f := range dto.Fields {
		isPointer := strings.HasPrefix(f.Type, "*")
		baseType := strings.TrimPrefix(f.Type, "*")

		var condition string
		if !f.Optional {
			condition = ""
		} else if isPointer {
			condition = fmt.Sprintf("if (%s->%s) ", varName, f.Name)
		} else {
			// Non-pointer optional fields shouldn't exist for JSON payloads
			condition = ""
		}

		var fieldCode string
		switch baseType {
		case "string":
			if firstField {
				fieldCode = fmt.Sprintf("%ssnprintf(payload, sizeof payload, \"{\\\"%s\\\":\\\"%%s\\\"\", *%s->%s);",
					condition, f.JSONName, varName, f.Name)
			} else {
				fieldCode = fmt.Sprintf("%s{ char tmp[64]; snprintf(tmp, sizeof tmp, \",\\\"%s\\\":\\\"%%s\\\"\", *%s->%s); strncat_s(payload, sizeof(payload), tmp, sizeof(payload) - strlen(payload) - 1); }",
					condition, f.JSONName, varName, f.Name)
			}
		case "uint16", "uint32":
			if firstField {
				fieldCode = fmt.Sprintf("%ssnprintf(payload, sizeof payload, \"{\\\"%s\\\":%%u\", (unsigned)*%s->%s);",
					condition, f.JSONName, varName, f.Name)
			} else {
				fieldCode = fmt.Sprintf("%s{ char tmp[64]; snprintf(tmp, sizeof tmp, \",\\\"%s\\\":%%u\", (unsigned)*%s->%s); strncat_s(payload, sizeof(payload), tmp, sizeof(payload) - strlen(payload) - 1); }",
					condition, f.JSONName, varName, f.Name)
			}
		default:
			// Unsupported type, skip
			continue
		}

		if firstField {
			firstField = false
		}

		lines = append(lines, "        "+fieldCode)
	}

	// Close JSON object
	lines = append(lines, "        strncat_s(payload, sizeof(payload), \"}\", sizeof(payload) - strlen(payload) - 1);")
	lines = append(lines, "    }")

	return strings.Join(lines, "\n")
}

func cType(goType, kind string) string {
	switch {
	case strings.HasPrefix(goType, "[]"):
		elem := strings.TrimPrefix(goType, "[]")
		return cType(elem, "")
	}

	switch goType {
	case "string":
		return "const char*"
	case "uint8":
		return "uint8_t"
	case "uint16":
		return "uint16_t"
	case "uint32":
		return "uint32_t"
	case "uint64":
		return "uint64_t"
	case "int8":
		return "int8_t"
	case "int16":
		return "int16_t"
	case "int32":
		return "int32_t"
	case "int64":
		return "int64_t"
	case "bool":
		return "int"
	case "float32":
		return "float"
	case "float64":
		return "double"
	default:
		if kind == "struct" {
			return fmt.Sprintf("viiper_%s_t*", common.ToSnakeCase(goType))
		}
		return goType
	}
}

func hasWireTag(md *meta.Metadata, device, direction string) bool {
	if md.WireTags == nil {
		return false
	}
	return md.WireTags.HasDirection(device, direction)
}

func wireFields(md *meta.Metadata, device, direction string) string {
	if md.WireTags == nil {
		return "/* no wire tags found */"
	}

	tag := md.WireTags.GetTag(device, direction)
	if tag == nil {
		return "/* no wire tag for this device/direction */"
	}

	return renderCWireFields(tag)
}

func renderCWireFields(tag *scanner.WireTag) string {
	if tag == nil || len(tag.Fields) == 0 {
		return "/* no fields */"
	}

	var lines []string
	for _, field := range tag.Fields {
		if strings.Contains(field.Type, "*") {
			base := strings.Split(field.Type, "*")[0]
			cbase := wireBaseToC(base)
			lines = append(lines, fmt.Sprintf("%s* %s; size_t %s_count;", cbase, field.Name, common.ToSnakeCase(field.Name)))
			continue
		}
		lines = append(lines, fmt.Sprintf("%s %s;", wireBaseToC(field.Type), field.Name))
	}
	return strings.Join(lines, "\n    ")
}

func wireBaseToC(wireType string) string {
	switch wireType {
	case "u8":
		return "uint8_t"
	case "u16":
		return "uint16_t"
	case "u32":
		return "uint32_t"
	case "u64":
		return "uint64_t"
	case "i8":
		return "int8_t"
	case "i16":
		return "int16_t"
	case "i32":
		return "int32_t"
	case "i64":
		return "int64_t"
	default:
		return wireType
	}
}

func indent(spaces int, s string) string {
	prefix := strings.Repeat(" ", spaces)
	parts := strings.Split(s, "\n")
	for i, p := range parts {
		if p != "" {
			parts[i] = prefix + p
		}
	}
	return strings.Join(parts, "\n")
}

func orderedPathParams(path string) []string {
	if path == "" {
		return nil
	}
	parts := strings.Split(path, "/")
	var params []string
	for _, p := range parts {
		if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") {
			params = append(params, p[1:len(p)-1])
		}
	}
	return params
}

func fieldDecl(md *meta.Metadata, f scanner.FieldInfo) string {
	if f.TypeKind == "slice" || strings.HasPrefix(f.Type, "[]") {
		elem := strings.TrimPrefix(f.Type, "[]")
		cElem := fieldTypeToCType(md, elem)
		return fmt.Sprintf("%s* %s; size_t %s_count;%s", cElem, f.Name, common.ToSnakeCase(f.Name), optComment(f))
	}

	if strings.HasPrefix(f.Type, "*") {
		elem := strings.TrimPrefix(f.Type, "*")
		cElem := fieldTypeToCType(md, elem)
		return fmt.Sprintf("%s* %s;%s", cElem, f.Name, optComment(f))
	}

	return fmt.Sprintf("%s %s;%s", fieldTypeToCType(md, f.Type), f.Name, optComment(f))
}

// fieldTypeToCType converts a field type to C type, using metadata for struct type name mapping
func fieldTypeToCType(md *meta.Metadata, goType string) string {
	// Check if it's a primitive type first
	switch goType {
	case "string":
		return "const char*"
	case "uint8":
		return "uint8_t"
	case "uint16":
		return "uint16_t"
	case "uint32":
		return "uint32_t"
	case "uint64":
		return "uint64_t"
	case "int8":
		return "int8_t"
	case "int16":
		return "int16_t"
	case "int32":
		return "int32_t"
	case "int", "int64":
		return "int64_t"
	case "bool":
		return "int"
	case "float32":
		return "float"
	case "float64":
		return "double"
	}

	// For struct types, check if it's a DTO and use the type name mapping
	for _, dto := range md.DTOs {
		if dto.Name == goType {
			return fmt.Sprintf("viiper_%s_t", dtoToCTypeName(md, goType))
		}
	}

	// Default: assume it's a struct and use snake_case
	return fmt.Sprintf("viiper_%s_t", common.ToSnakeCase(goType))
}

func optComment(f scanner.FieldInfo) string {
	if f.Optional {
		return " /* optional */"
	}
	return ""
}

func mapFuncDecl(device string, mapInfo scanner.MapInfo) string {
	keyType := mapGoTypeToCType(mapInfo.KeyType)
	valueType := mapGoTypeToCType(mapInfo.ValueType)
	funcName := fmt.Sprintf("viiper_%s_%s_lookup", device, common.ToSnakeCase(mapInfo.Name))

	return fmt.Sprintf("int %s(%s key, %s* out_value);", funcName, keyType, valueType)
}

func mapFuncImpl(device string, mapInfo scanner.MapInfo) string {
	keyType := mapGoTypeToCType(mapInfo.KeyType)
	valueType := mapGoTypeToCType(mapInfo.ValueType)
	funcName := fmt.Sprintf("viiper_%s_%s_lookup", device, common.ToSnakeCase(mapInfo.Name))

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("int %s(%s key, %s* out_value) {\n", funcName, keyType, valueType))
	builder.WriteString("    if (!out_value) return 0;\n")
	builder.WriteString("    switch (key) {\n")

	keys := common.SortedStringKeys(mapInfo.Entries)
	for _, keyStr := range keys {
		value := mapInfo.Entries[keyStr]
		cKey := formatCMapKey(keyStr, mapInfo.KeyType, device)
		cValue := formatCMapValue(value, mapInfo.ValueType, device)
		builder.WriteString(fmt.Sprintf("        case %s: *out_value = %s; return 1;\n", cKey, cValue))
	}

	builder.WriteString("        default: return 0;\n")
	builder.WriteString("    }\n")
	builder.WriteString("}\n")

	return builder.String()
}

func mapGoTypeToCType(goType string) string {
	switch goType {
	case "byte", "uint8":
		return "uint8_t"
	case "uint16":
		return "uint16_t"
	case "uint32", "uint":
		return "uint32_t"
	case "uint64":
		return "uint64_t"
	case "int8":
		return "int8_t"
	case "int16":
		return "int16_t"
	case "int32", "int":
		return "int32_t"
	case "int64":
		return "int64_t"
	case "bool":
		return "int"
	case "string":
		return "const char*"
	default:
		return goType
	}
}

func formatCMapKey(key string, goType string, device string) string {
	switch goType {
	case "byte", "uint8":
		if len(key) == 1 {
			// Escape special characters
			switch key[0] {
			case '\n':
				return "'\\n'"
			case '\r':
				return "'\\r'"
			case '\t':
				return "'\\t'"
			case '\\':
				return "'\\\\'"
			case '\'':
				return "'\\''"
			case 0:
				return "'\\0'"
			}
			if key[0] >= 32 && key[0] <= 126 {
				return fmt.Sprintf("'%s'", key)
			}
			return fmt.Sprintf("0x%02X", key[0])
		}
		if len(key) > 0 && (key[0] >= 'A' && key[0] <= 'Z') {
			prefix := common.ExtractPrefix(key)
			if prefix != "" {
				constName := strings.ToUpper(key)
				return fmt.Sprintf("VIIPER_%s_%s", strings.ToUpper(device), constName)
			}
		}
		return key
	case "string":
		return fmt.Sprintf("\"%s\"", key)
	default:
		return key
	}
}

func formatCMapValue(value interface{}, goType string, device string) string {
	switch goType {
	case "byte", "uint8", "uint16", "uint32", "uint64":
		if str, ok := value.(string); ok && !strings.Contains(str, " ") {
			if len(str) > 0 && (str[0] >= 'A' && str[0] <= 'Z') {
				prefix := common.ExtractPrefix(str)
				if prefix != "" {
					constName := strings.ToUpper(str)
					return fmt.Sprintf("VIIPER_%s_%s", strings.ToUpper(device), constName)
				}
			}
			return str
		}
		if num, ok := value.(int64); ok {
			return fmt.Sprintf("0x%X", num)
		}
		if num, ok := value.(uint64); ok {
			return fmt.Sprintf("0x%X", num)
		}
		return fmt.Sprintf("%v", value)
	case "bool":
		if b, ok := value.(bool); ok {
			if b {
				return "1"
			}
			return "0"
		}
		if str, ok := value.(string); ok {
			if str == "true" {
				return "1"
			}
			if str == "false" {
				return "0"
			}
			return str
		}
		return "0"
	case "string":
		if str, ok := value.(string); ok {
			return fmt.Sprintf("\"%s\"", str)
		}
		return fmt.Sprintf("\"%v\"", value)
	default:
		return fmt.Sprintf("%v", value)
	}
}

// generateFreeFunction creates a free function for a DTO
func generateFreeFunction(md *meta.Metadata, dto scanner.DTOSchema) string {
	snakeName := dtoToCTypeName(md, dto.Name)
	lines := []string{
		fmt.Sprintf("VIIPER_API void viiper_free_%s(viiper_%s_t* v){", snakeName, snakeName),
	}

	hasPointers := false
	for _, f := range dto.Fields {
		if strings.HasPrefix(f.Type, "*") || strings.HasPrefix(f.Type, "[]") {
			hasPointers = true
			break
		}
	}

	if !hasPointers {
		lines = append(lines, " (void)v; }")
		return strings.Join(lines, "")
	}

	lines = append(lines, " if (!v) return;")

	for _, f := range dto.Fields {
		baseType := strings.TrimPrefix(f.Type, "*")
		baseType = strings.TrimPrefix(baseType, "[]")

		if strings.HasPrefix(f.Type, "[]") {
			// Array type - need to free elements and array itself
			if baseType != "string" && baseType != "uint8" && baseType != "uint16" && baseType != "uint32" && baseType != "uint64" {
				// Complex type array - need to find the DTO and free each field
				lines = append(lines, fmt.Sprintf(" if (v->%s){ for (size_t i=0;i<v->%s_count;i++){ ", f.Name, common.ToSnakeCase(f.JSONName)))
				// Look up the DTO to find string fields that need freeing
				for _, dtoSchema := range md.DTOs {
					if dtoSchema.Name == baseType {
						for _, field := range dtoSchema.Fields {
							fieldType := strings.TrimPrefix(field.Type, "*")
							if fieldType == "string" {
								lines = append(lines, fmt.Sprintf("if (v->%s[i].%s) free((void*)v->%s[i].%s); ", f.Name, field.Name, f.Name, field.Name))
							}
						}
						break
					}
				}
				lines = append(lines, " } free(v->"+f.Name+");}")
			} else if baseType == "string" {
				lines = append(lines, fmt.Sprintf(" if (v->%s){ for (size_t i=0;i<v->%s_count;i++){ if (v->%s[i]) free((void*)v->%s[i]); } free(v->%s);}", f.Name, common.ToSnakeCase(f.JSONName), f.Name, f.Name, f.Name))
			} else {
				// Primitive array - just free the array
				lines = append(lines, fmt.Sprintf(" if (v->%s) free(v->%s);", f.Name, f.Name))
			}
		} else if strings.HasPrefix(f.Type, "*") && baseType == "string" {
			// Pointer to string
			lines = append(lines, fmt.Sprintf(" if (v->%s) free((void*)v->%s);", f.Name, f.Name))
		}
	}

	lines = append(lines, " }")
	return strings.Join(lines, "")
}

// generateParser creates a parser function for a DTO
func generateParser(md *meta.Metadata, dto scanner.DTOSchema) string {
	snakeName := dtoToCTypeName(md, dto.Name)

	// For parsing object responses, use _obj suffix
	parserName := fmt.Sprintf("viiper_parse_%s", snakeName)
	if strings.HasSuffix(snakeName, "_info") {
		parserName = fmt.Sprintf("viiper_parse_%s_obj", snakeName)
	}

	lines := []string{
		fmt.Sprintf("static int %s(const char* json, viiper_%s_t* out){", parserName, snakeName),
	}

	var parserCalls []string
	for _, f := range dto.Fields {
		baseType := strings.TrimPrefix(f.Type, "*")
		baseType = strings.TrimPrefix(baseType, "[]")

		required := !f.Optional

		if strings.HasPrefix(f.Type, "[]") {
			if baseType == "uint32" {
				parserCalls = append(parserCalls, fmt.Sprintf("json_parse_array_uint32(json, \"%s\", &out->%s, &out->%s_count)", f.JSONName, f.Name, common.ToSnakeCase(f.JSONName)))
			} else {
				// Complex type - use appropriate parser based on type name mapping
				parserFuncName := fmt.Sprintf("json_parse_array_%s", dtoToCTypeName(md, baseType))
				parserCalls = append(parserCalls, fmt.Sprintf("%s(json, \"%s\", &out->%s, &out->%s_count)", parserFuncName, f.JSONName, f.Name, common.ToSnakeCase(f.JSONName)))
			}
		} else if baseType == "string" {
			if required {
				parserCalls = append(parserCalls, fmt.Sprintf("json_parse_string_alloc(json, \"%s\", (char**)&out->%s)", f.JSONName, f.Name))
			} else {
				// Optional string - don't fail if missing
				lines = append(lines, fmt.Sprintf(" json_parse_string_alloc(json, \"%s\", (char**)&out->%s);", f.JSONName, f.Name))
			}
		} else if baseType == "uint32" {
			parserCalls = append(parserCalls, fmt.Sprintf("json_parse_uint32(json, \"%s\", &out->%s)", f.JSONName, f.Name))
		}
	}

	if len(parserCalls) == 0 {
		lines = append(lines, " return 0; }")
	} else if len(parserCalls) == 1 {
		lines = append(lines, fmt.Sprintf(" return %s==0?0:-1; }", parserCalls[0]))
	} else {
		for i, call := range parserCalls {
			if i < len(parserCalls)-1 {
				lines = append(lines, fmt.Sprintf(" if (%s!=0) return -1;", call))
			} else {
				lines = append(lines, fmt.Sprintf(" return %s==0?0:-1; }", call))
			}
		}
	}

	return strings.Join(lines, "")
}
