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
		HasArray  bool
		ArrayInfo *arrayFieldInfo
	}{
		Device:    device,
		ClassName: className,
	}

	for _, field := range tag.Fields {
		wf := wireField{
			Name:   toPascalCase(field.Name),
			GoType: field.Type,
			CSType: mapGoTypeToCSharp(field.Type),
		}

		if strings.Contains(field.Spec, "*") {
			parts := strings.Split(field.Spec, "*")
			if len(parts) == 2 {
				data.HasArray = true
				data.ArrayInfo = &arrayFieldInfo{
					FieldName:      wf.Name,
					CountFieldName: toPascalCase(parts[1]),
					ElementType:    wf.CSType,
				}
				wf.IsArray = true
			}
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
	Name    string
	GoType  string
	CSType  string
	IsArray bool
}

type arrayFieldInfo struct {
	FieldName      string
	CountFieldName string
	ElementType    string
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
{{range .Fields}}    public required {{.CSType}}{{if .IsArray}}[]{{end}} {{.Name}} { get; set; }
{{end}}
    public void Write(BinaryWriter writer)
    {
{{if .HasArray}}        // Write fixed fields
{{range .Fields}}{{if not .IsArray}}        writer.Write({{.Name}});
{{end}}{{end}}
        // Write variable-length array
        for (int i = 0; i < {{.ArrayInfo.CountFieldName}}; i++)
        {
            writer.Write({{.ArrayInfo.FieldName}}[i]);
        }
{{else}}{{range .Fields}}        writer.Write({{.Name}});
{{end}}{{end}}    }

    /// <summary>
    /// Read from binary stream (for receiving output from server).
    /// </summary>
    public static {{.Device}}{{.ClassName}} Read(BinaryReader reader)
    {
{{if .HasArray}}        // Read fixed fields
{{range .Fields}}{{if not .IsArray}}        var {{toCamel .Name}} = reader.Read{{readerMethod .CSType}}();
{{end}}{{end}}
        // Read variable-length array
        var {{toCamel .ArrayInfo.FieldName}} = new {{.ArrayInfo.ElementType}}[{{toCamel .ArrayInfo.CountFieldName}}];
        for (int i = 0; i < {{toCamel .ArrayInfo.CountFieldName}}; i++)
        {
            {{toCamel .ArrayInfo.FieldName}}[i] = reader.Read{{readerMethod .ArrayInfo.ElementType}}();
        }
        
        return new {{.Device}}{{.ClassName}}
        {
{{range .Fields}}{{if not .IsArray}}            {{.Name}} = {{toCamel .Name}},
{{end}}{{end}}            {{.ArrayInfo.FieldName}} = {{toCamel .ArrayInfo.FieldName}}
        };
{{else}}        return new {{.Device}}{{.ClassName}}
        {
{{range .Fields}}            {{.Name}} = reader.Read{{readerMethod .CSType}}(),
{{end}}        };
{{end}}    }
}
`
