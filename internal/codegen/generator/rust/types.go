package rust

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/Alia5/VIIPER/internal/codegen/common"
	"github.com/Alia5/VIIPER/internal/codegen/meta"
)

const typesTemplate = `{{.Header}}
// Management API DTO types

{{range .DTOs}}
/// {{.Name}} DTO
#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct {{.Name}} {
{{- range .Fields}}
    {{if .Optional}}#[serde(skip_serializing_if = "Option::is_none")]
    {{end}}{{if ne .JSONName .RustName}}#[serde(rename = "{{.JSONName}}")]
    {{end}}pub {{.RustName}}: {{.RustType}},
{{end}}}

{{end}}
`

type rustFieldInfo struct {
	Name     string
	JSONName string
	RustName string
	RustType string
	Optional bool
}

type rustDTOInfo struct {
	Name   string
	Fields []rustFieldInfo
}

func generateTypes(logger *slog.Logger, srcDir string, md *meta.Metadata) error {
	logger.Debug("Generating types.rs (management API DTOs)")
	outputFile := filepath.Join(srcDir, "types.rs")

	var dtos []rustDTOInfo
	for _, dto := range md.DTOs {
		rustDTO := rustDTOInfo{
			Name:   dto.Name,
			Fields: []rustFieldInfo{},
		}

		for _, field := range dto.Fields {
			rustName := common.ToSnakeCase(field.Name)
			if isRustKeyword(rustName) {
				rustName = "r#" + rustName
			}
			rustType := fieldTypeToRust(field)

			rustDTO.Fields = append(rustDTO.Fields, rustFieldInfo{
				Name:     field.Name,
				JSONName: field.JSONName,
				RustName: rustName,
				RustType: rustType,
				Optional: field.Optional,
			})
		}

		dtos = append(dtos, rustDTO)
	}

	funcMap := template.FuncMap{}
	tmpl, err := template.New("types").Funcs(funcMap).Parse(typesTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	data := struct {
		Header string
		DTOs   []rustDTOInfo
	}{
		Header: writeFileHeaderRust(),
		DTOs:   dtos,
	}

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	logger.Info("Generated types.rs", "file", outputFile)
	return nil
}

func fieldTypeToRust(field interface{}) string {
	v := reflect.ValueOf(field)
	typeStr := v.FieldByName("Type").String()
	typeKind := v.FieldByName("TypeKind").String()
	optional := v.FieldByName("Optional").Bool()

	var rustType string

	if typeKind == "map" || strings.HasPrefix(typeStr, "map[") {
		keyType, valueType, ok := parseGoMapType(typeStr)
		if ok {
			_ = keyType
			rustType = fmt.Sprintf("std::collections::HashMap<String, %s>", goTypeToRust(valueType))
		} else {
			rustType = "std::collections::HashMap<String, serde_json::Value>"
		}
	} else if typeKind == "slice" || strings.HasPrefix(typeStr, "[]") {
		elem := strings.TrimPrefix(typeStr, "[]")
		rustType = fmt.Sprintf("Vec<%s>", goTypeToRust(elem))
	} else {
		rustType = goTypeToRust(typeStr)
	}

	if optional && !strings.HasPrefix(rustType, "Option<") {
		rustType = fmt.Sprintf("Option<%s>", rustType)
	}

	return rustType
}

func isRustKeyword(s string) bool {
	keywords := map[string]bool{
		"as": true, "break": true, "const": true, "continue": true, "crate": true,
		"else": true, "enum": true, "extern": true, "false": true, "fn": true,
		"for": true, "if": true, "impl": true, "in": true, "let": true,
		"loop": true, "match": true, "mod": true, "move": true, "mut": true,
		"pub": true, "ref": true, "return": true, "self": true, "Self": true,
		"static": true, "struct": true, "super": true, "trait": true, "true": true,
		"type": true, "unsafe": true, "use": true, "where": true, "while": true,
		"async": true, "await": true, "dyn": true,
	}
	return keywords[s]
}
