package csharp

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/Alia5/VIIPER/internal/codegen/meta"
)

const dtoTemplate = `{{writeFileHeader}}using System.Collections.Generic;
using System.Text.Json.Serialization;

namespace Viiper.Client.Types;

{{range .DTOs}}
/// <summary>
/// {{.Name}} DTO
/// </summary>
public class {{.Name}}
{
{{- range .Fields}}
    [JsonPropertyName("{{.JSONName}}")]
    public {{if not .Optional}}required {{end}}{{fieldTypeToCSharp .}}{{if .Optional}}?{{end}} {{.Name}} { get; set; }
{{end}}
}

{{end}}
`

func generateTypes(logger *slog.Logger, typesDir string, md *meta.Metadata) error {
	logger.Debug("Generating management API DTO types")

	outputFile := filepath.Join(typesDir, "ManagementDtos.cs")

	funcMap := template.FuncMap{
		"toPascalCase":      toPascalCase,
		"fieldTypeToCSharp": fieldTypeToCSharp,
		"writeFileHeader":   writeFileHeader,
	}

	tmpl, err := template.New("dtos").Funcs(funcMap).Parse(dtoTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	data := struct {
		DTOs interface{}
	}{
		DTOs: md.DTOs,
	}

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	logger.Info("Generated DTO types", "file", outputFile)
	return nil
}

func fieldTypeToCSharp(field interface{}) string {
	v := reflect.ValueOf(field)
	typeStr := v.FieldByName("Type").String()
	typeKind := v.FieldByName("TypeKind").String()

	if typeKind == "map" || strings.HasPrefix(typeStr, "map[") {
		keyType, valueType, ok := parseGoMapType(typeStr)
		if !ok {
			return "Dictionary<string, object>"
		}
		_ = keyType
		csVal := goTypeToCSharp(valueType)
		if valueType == "any" || valueType == "interface{}" {
			csVal = "object?"
		}
		return "Dictionary<string, " + csVal + ">"
	}

	if typeKind == "slice" || strings.HasPrefix(typeStr, "[]") {
		elemType := strings.TrimPrefix(typeStr, "[]")
		csharpElemType := goTypeToCSharp(elemType)
		return csharpElemType + "[]"
	}

	if typeKind == "struct" {
		return toPascalCase(typeStr)
	}

	return goTypeToCSharp(typeStr)
}
