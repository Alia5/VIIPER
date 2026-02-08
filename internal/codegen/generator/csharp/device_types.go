package csharp

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/Alia5/VIIPER/internal/codegen/meta"
	"github.com/Alia5/VIIPER/internal/codegen/scanner"
)

func generateDeviceTypes(logger *slog.Logger, deviceDir string, deviceName string, md *meta.Metadata) error {
	logger.Debug("Generating device types", "device", deviceName)

	if md.WireTags == nil {
		logger.Warn("No wire tags metadata available")
		return nil
	}

	c2sTag := md.WireTags.GetTag(deviceName, "c2s")
	s2cTag := md.WireTags.GetTag(deviceName, "s2c")

	if c2sTag == nil && s2cTag == nil {
		logger.Warn("No wire tags found for device", "device", deviceName)
		return nil
	}

	pascalDevice := toPascalCase(deviceName)

	if c2sTag != nil {
		inputPath := filepath.Join(deviceDir, pascalDevice+"Input.cs")
		if err := generateWireClass(inputPath, pascalDevice, "Input", c2sTag); err != nil {
			return fmt.Errorf("generating Input: %w", err)
		}
		logger.Debug("Generated Input class", "device", deviceName, "path", inputPath)
	}

	if s2cTag != nil {
		outputPath := filepath.Join(deviceDir, pascalDevice+"Output.cs")
		if err := generateWireClass(outputPath, pascalDevice, "Output", s2cTag); err != nil {
			return fmt.Errorf("generating Output: %w", err)
		}
		logger.Debug("Generated Output class", "device", deviceName, "path", outputPath)
	}

	logger.Info("Generated device types", "device", deviceName)
	return nil
}

func generateWireClass(outputPath, device, className string, tag *scanner.WireTag) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	data := struct {
		Device    string
		ClassName string
		Fields    []wireField
	}{
		Device:    device,
		ClassName: className,
	}

	for _, field := range tag.Fields {
		wf := wireField{Name: toPascalCase(field.Name)}
		if idx := strings.Index(field.Type, "*"); idx >= 0 {
			wf.IsArray = true
			baseType := field.Type[:idx]
			wf.CSType = mapGoTypeToCSharp(baseType)
			countToken := field.Type[idx+1:]
			if n, err := strconv.Atoi(countToken); err == nil {
				wf.FixedLen = n
			} else {
				wf.CountFieldName = toPascalCase(countToken)
			}
		} else {
			wf.CSType = mapGoTypeToCSharp(field.Type)
		}

		data.Fields = append(data.Fields, wf)
	}

	funcMap := template.FuncMap{
		"readerMethod": getCSharpReaderMethod,
		"toCamel":      toCamelCase,
	}

	tmpl := template.Must(template.New("wireclass").Funcs(funcMap).Parse(wireClassTemplate))
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	return nil
}

type wireField struct {
	Name           string
	CSType         string
	IsArray        bool
	CountFieldName string
	FixedLen       int
}

func mapGoTypeToCSharp(goType string) string {
	switch goType {
	case "u8":
		return "byte"
	case "u16":
		return "ushort"
	case "u32":
		return "uint"
	case "u64":
		return "ulong"
	case "i8":
		return "sbyte"
	case "i16":
		return "short"
	case "i32":
		return "int"
	case "i64":
		return "long"
	default:
		return "byte"
	}
}

func getCSharpReaderMethod(csType string) string {
	switch csType {
	case "byte":
		return "Byte"
	case "sbyte":
		return "SByte"
	case "ushort":
		return "UInt16"
	case "short":
		return "Int16"
	case "uint":
		return "UInt32"
	case "int":
		return "Int32"
	case "ulong":
		return "UInt64"
	case "long":
		return "Int64"
	default:
		return "Byte"
	}
}

const wireClassTemplate = `using System;
using System.IO;

namespace Viiper.Client.Devices.{{.Device}};

/// <summary>
/// Wire protocol {{.ClassName}} message for {{.Device}} device.
/// </summary>
public class {{.Device}}{{.ClassName}} : IBinarySerializable
{
{{range .Fields}}{{if and .IsArray (gt .FixedLen 0)}}    public {{.CSType}}[] {{.Name}} { get; set; } = new {{.CSType}}[{{.FixedLen}}];
{{else}}    public required {{.CSType}}{{if .IsArray}}[]{{end}} {{.Name}} { get; set; }
{{end}}{{end}}
    public void Write(BinaryWriter writer)
    {
{{range .Fields}}{{if .IsArray}}{{if gt .FixedLen 0}}        for (int i = 0; i < {{.FixedLen}}; i++)
		{
			writer.Write(({{.Name}} != null && i < {{.Name}}.Length) ? {{.Name}}[i] : default({{.CSType}}));
		}
{{else}}        for (int i = 0; i < {{.CountFieldName}}; i++)
		{
			writer.Write({{.Name}}[i]);
		}
{{end}}{{else}}        writer.Write({{.Name}});
{{end}}{{end}}    }

    /// <summary>
    /// Read from binary stream (for receiving output from server).
    /// </summary>
    public static {{.Device}}{{.ClassName}} Read(BinaryReader reader)
    {
	{{range .Fields}}{{if .IsArray}}{{if gt .FixedLen 0}}        var {{toCamel .Name}} = new {{.CSType}}[{{.FixedLen}}];
		for (int i = 0; i < {{.FixedLen}}; i++)
		{
		    {{toCamel .Name}}[i] = reader.Read{{readerMethod .CSType}}();
		}
	{{else}}        var {{toCamel .Name}} = new {{.CSType}}[{{toCamel .CountFieldName}}];
		for (int i = 0; i < {{toCamel .CountFieldName}}; i++)
		{
		    {{toCamel .Name}}[i] = reader.Read{{readerMethod .CSType}}();
		}
	{{end}}{{else}}        var {{toCamel .Name}} = reader.Read{{readerMethod .CSType}}();
	{{end}}{{end}}

		return new {{.Device}}{{.ClassName}}
		{
	{{range .Fields}}            {{.Name}} = {{toCamel .Name}},
	{{end}}        };
	    }
}
`
