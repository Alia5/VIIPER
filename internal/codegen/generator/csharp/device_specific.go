package csharp

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Alia5/VIIPER/internal/codegen/meta"
	"github.com/Alia5/VIIPER/internal/codegen/scanner"
)

type csDeviceSpecificField struct {
	Name     string
	JSONName string
	CSType   string
	Optional bool
}

type csDeviceSpecificStruct struct {
	Name   string
	Fields []csDeviceSpecificField
}

const deviceSpecificTemplate = `{{writeFileHeader}}using System;
using System.Collections.Generic;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace Viiper.Client.Devices.{{.Device}};

{{range .Structs}}
public class {{.Name}}
{
{{- range .Fields}}
    [JsonPropertyName("{{.JSONName}}")]
    public {{.CSType}}{{if .Optional}}?{{end}} {{.Name}} { get; set; }{{if and (not .Optional) (eq .CSType "string")}} = string.Empty;{{end}}
{{end}}

    public Dictionary<string, object?> ToMap()
    {
        var json = JsonSerializer.Serialize(this);
        return JsonSerializer.Deserialize<Dictionary<string, object?>>(json) ?? new Dictionary<string, object?>();
    }

    public static {{.Name}} FromMap(Dictionary<string, object?> map)
    {
        var json = JsonSerializer.Serialize(map);
        return JsonSerializer.Deserialize<{{.Name}}>(json) ?? new {{.Name}}();
    }
}

{{end}}
`

func generateDeviceSpecific(logger *slog.Logger, deviceDir string, deviceName string, md *meta.Metadata) error {
	structs := md.DeviceStructs[deviceName]
	if len(structs) == 0 {
		return nil
	}

	logger.Debug("Generating C# meta helpers", "device", deviceName)

	pascalDevice := toPascalCase(deviceName)
	legacyPath := filepath.Join(deviceDir, pascalDevice+"DeviceSpecific.cs")
	if err := os.Remove(legacyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove legacy C# device-specific file: %w", err)
	}
	outputPath := filepath.Join(deviceDir, pascalDevice+"Meta.cs")

	data := struct {
		Device  string
		Structs []csDeviceSpecificStruct
	}{
		Device:  pascalDevice,
		Structs: make([]csDeviceSpecificStruct, 0, len(structs)),
	}

	for _, s := range structs {
		entry := csDeviceSpecificStruct{
			Name:   s.Name,
			Fields: make([]csDeviceSpecificField, 0, len(s.Fields)),
		}
		for _, f := range s.Fields {
			entry.Fields = append(entry.Fields, csDeviceSpecificField{
				Name:     f.Name,
				JSONName: f.JSONName,
				CSType:   fieldTypeToCSharpForDeviceSpecific(f),
				Optional: f.Optional,
			})
		}
		data.Structs = append(data.Structs, entry)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close() //nolint:errcheck

	tmpl := template.Must(template.New("deviceSpecificCS").Funcs(template.FuncMap{
		"writeFileHeader": writeFileHeader,
	}).Parse(deviceSpecificTemplate))

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	logger.Info("Generated C# meta helpers", "device", deviceName, "path", outputPath)
	return nil
}

func fieldTypeToCSharpForDeviceSpecific(field scanner.FieldInfo) string {
	typeStr := field.Type
	typeKind := field.TypeKind

	if typeKind == "map" || strings.HasPrefix(typeStr, "map[") {
		valueType, ok := parseGoMapType(typeStr)
		if !ok {
			return "Dictionary<string, object?>"
		}
		csVal := goTypeToCSharp(valueType)
		if valueType == "any" || valueType == "interface{}" {
			csVal = "object?"
		}
		return "Dictionary<string, " + csVal + ">"
	}

	if typeKind == "slice" || strings.HasPrefix(typeStr, "[]") {
		elemType := strings.TrimPrefix(typeStr, "[]")
		return goTypeToCSharp(elemType) + "[]"
	}

	if typeStr == "time.Time" {
		return "DateTime"
	}

	if typeKind == "struct" {
		return toPascalCase(typeStr)
	}

	return goTypeToCSharp(typeStr)
}
