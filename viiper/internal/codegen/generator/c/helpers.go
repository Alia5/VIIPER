package cgen

import (
	"fmt"
	"strings"
	"text/template"
	"viiper/internal/codegen/meta"
	"viiper/internal/codegen/scanner"
)

func tplFuncs(md *meta.Metadata) template.FuncMap {
	return template.FuncMap{
		"ctype":       cType,
		"snakecase":   toSnakeCase,
		"upper":       strings.ToUpper,
		"hasWireTag":  func(device, direction string) bool { return hasWireTag(md, device, direction) },
		"wireFields":  func(device, direction string) string { return wireFields(md, device, direction) },
		"indent":      indent,
		"hasResponse": hasResponse,
		"fieldDecl":   fieldDecl,
		"pathParams":  orderedPathParams,
		"join":        strings.Join,
	}
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
		if goType == "Device" {
			return "viiper_device_info_t"
		}
		if kind == "struct" {
			return fmt.Sprintf("viiper_%s_t*", toSnakeCase(goType))
		}
		return goType
	}
}

func toSnakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
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
			lines = append(lines, fmt.Sprintf("%s* %s; size_t %s_count;", cbase, field.Name, toSnake(field.Name)))
			continue
		}
		lines = append(lines, fmt.Sprintf("%s %s;", wireBaseToC(field.Type), field.Name))
	}
	return strings.Join(lines, "\n    ")
}

func toSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
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

func hasResponse(handler string) bool { return false }

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

func fieldDecl(f scanner.FieldInfo) string {
	if f.TypeKind == "slice" || strings.HasPrefix(f.Type, "[]") {
		elem := strings.TrimPrefix(f.Type, "[]")
		cElem := cType(elem, "")
		return fmt.Sprintf("%s* %s; size_t %s_count;%s", cElem, f.Name, toSnakeCase(f.Name), optComment(f))
	}
	return fmt.Sprintf("%s %s;%s", cType(f.Type, f.TypeKind), f.Name, optComment(f))
}

func optComment(f scanner.FieldInfo) string {
	if f.Optional {
		return " /* optional */"
	}
	return ""
}
