package typescript

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

const dtoTemplateTS = `{{writeFileHeaderTS}}
// Management API DTO interfaces

{{range .DTOs}}
// {{.Name}} DTO
export interface {{.Name}} {
{{- range .Fields}}
  {{.JSONName}}{{if .Optional}}?{{end}}: {{fieldTypeToTS .}};
{{end}}
}

{{end}}
`

func generateTypes(logger *slog.Logger, typesDir string, md *meta.Metadata) error {
	logger.Debug("Generating management API DTO interfaces (TypeScript)")
	outputFile := filepath.Join(typesDir, "ManagementDtos.ts")

	funcMap := template.FuncMap{
		"writeFileHeaderTS": writeFileHeaderTS,
		"fieldTypeToTS":     fieldTypeToTS,
	}
	tmpl, err := template.New("dtos").Funcs(funcMap).Parse(dtoTemplateTS)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()
	data := struct{ DTOs interface{} }{DTOs: md.DTOs}
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}
	logger.Info("Generated DTO interfaces", "file", outputFile)
	return nil
}

func fieldTypeToTS(field interface{}) string {
	v := reflect.ValueOf(field)
	typeStr := v.FieldByName("Type").String()
	typeKind := v.FieldByName("TypeKind").String()

	if typeKind == "map" || strings.HasPrefix(typeStr, "map[") {
		k, val, ok := parseGoMapType(typeStr)
		if ok {
			_ = k
			return "Record<string, " + goTypeToTS(val) + ">"
		}
		return "Record<string, unknown>"
	}

	if typeKind == "slice" || strings.HasPrefix(typeStr, "[]") {
		elem := strings.TrimPrefix(typeStr, "[]")
		return goTypeToTS(elem) + "[]"
	}
	if typeKind == "struct" {
		return common.ToPascalCase(typeStr)
	}
	return goTypeToTS(typeStr)
}
