package typescript

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"viiper/internal/codegen/common"
	"viiper/internal/codegen/meta"
	"viiper/internal/codegen/scanner"
)

func generateDeviceTypes(logger *slog.Logger, deviceDir string, deviceName string, md *meta.Metadata) error {
	logger.Debug("Generating TS device types", "device", deviceName)
	if md.WireTags == nil {
		return nil
	}
	c2sTag := md.WireTags.GetTag(deviceName, "c2s")
	s2cTag := md.WireTags.GetTag(deviceName, "s2c")
	pascalDevice := common.ToPascalCase(deviceName)
	if c2sTag != nil {
		path := filepath.Join(deviceDir, pascalDevice+"Input.ts")
		if err := generateWireClassTS(path, pascalDevice, "Input", c2sTag); err != nil {
			return err
		}
	}
	if s2cTag != nil {
		path := filepath.Join(deviceDir, pascalDevice+"Output.ts")
		if err := generateWireClassTS(path, pascalDevice, "Output", s2cTag); err != nil {
			return err
		}
	}
	return nil
}

type tsWireField struct {
	Name      string
	GoType    string
	Writer    string
	Reader    string
	IsArray   bool
	CountName string
}

func writerFor(goType string) string {
	switch goType {
	case "u8":
		return "writeU8"
	case "i8":
		return "writeI8"
	case "u16":
		return "writeU16LE"
	case "i16":
		return "writeI16LE"
	case "u32":
		return "writeU32LE"
	case "i32":
		return "writeI32LE"
	case "u64":
		return "writeU64LE"
	case "i64":
		return "writeI64LE"
	default:
		return "writeU8"
	}
}

func readerFor(goType string) string {
	switch goType {
	case "u8":
		return "readU8"
	case "i8":
		return "readI8"
	case "u16":
		return "readU16LE"
	case "i16":
		return "readI16LE"
	case "u32":
		return "readU32LE"
	case "i32":
		return "readI32LE"
	case "u64":
		return "readU64LE"
	case "i64":
		return "readI64LE"
	default:
		return "readU8"
	}
}

func generateWireClassTS(outputPath, device, className string, tag *scanner.WireTag) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()
	data := struct {
		Device    string
		ClassName string
		Fields    []tsWireField
		HasArray  bool
		ArrayInfo *tsWireField
	}{Device: device, ClassName: className}
	for _, field := range tag.Fields {
		wf := tsWireField{Name: common.ToPascalCase(field.Name), GoType: field.Type, Writer: writerFor(field.Type), Reader: readerFor(field.Type)}
		if strings.Contains(field.Spec, "*") {
			parts := strings.Split(field.Spec, "*")
			if len(parts) == 2 {
				data.HasArray = true
				wf.IsArray = true
				wf.CountName = common.ToPascalCase(parts[1])
				copy := wf
				data.ArrayInfo = &copy
			}
		}
		data.Fields = append(data.Fields, wf)
	}
	funcMap := template.FuncMap{
		"writeFileHeaderTS": writeFileHeaderTS,
		"toCamelTS": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToLower(s[:1]) + s[1:]
		},
	}
	tmpl := template.Must(template.New("wirets").Funcs(funcMap).Parse(wireClassTemplateTS))
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}
	return nil
}

const wireClassTemplateTS = `{{writeFileHeaderTS}}
import { BinaryWriter, BinaryReader } from '../../utils/binary';
import type { IBinarySerializable } from '../../ViiperDevice';

export class {{.Device}}{{.ClassName}} implements IBinarySerializable {
{{range .Fields}}  {{.Name}}!: {{if or (eq .GoType "u64") (eq .GoType "i64")}}bigint{{else}}number{{end}}{{if .IsArray}}[]{{end}};
{{end}}
  constructor(init: Partial<{{.Device}}{{.ClassName}}> = {}) {
    Object.assign(this, init);
  }
  write(writer: BinaryWriter): void {
{{if .HasArray}}    // Write fixed-size fields
{{range .Fields}}{{if not .IsArray}}    writer.{{.Writer}}(this.{{.Name}} as any);
{{end}}{{end}}    // Write variable-length array
    for (let i = 0; i < (this.{{.ArrayInfo.CountName}} as number); i++) {
      writer.{{.ArrayInfo.Writer}}((this.{{.ArrayInfo.Name}} as any[])[i]);
    }
{{else}}{{range .Fields}}    writer.{{.Writer}}(this.{{.Name}} as any);
{{end}}{{end}}  }
  static read(reader: BinaryReader): {{.Device}}{{.ClassName}} {
{{if .HasArray}}    // Read fixed-size fields
{{range .Fields}}{{if not .IsArray}}    const {{.Name | toCamelTS}} = reader.{{.Reader}}();
{{end}}{{end}}    const {{.ArrayInfo.Name | toCamelTS}}: any[] = new Array(Number({{.ArrayInfo.CountName | toCamelTS}}));
    for (let i = 0; i < Number({{.ArrayInfo.CountName | toCamelTS}}); i++) {
      {{.ArrayInfo.Name | toCamelTS}}[i] = reader.{{.ArrayInfo.Reader}}();
    }
    return new {{.Device}}{{.ClassName}}({
{{range .Fields}}{{if not .IsArray}}      {{.Name}}: {{.Name | toCamelTS}},
{{end}}{{end}}      {{.ArrayInfo.Name}}: {{.ArrayInfo.Name | toCamelTS}},
    });
{{else}}    return new {{.Device}}{{.ClassName}}({
{{range .Fields}}      {{.Name}}: reader.{{.Reader}}(),
{{end}}    });
{{end}}  }
}
`
