package typescript

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/Alia5/VIIPER/internal/codegen/common"
	"github.com/Alia5/VIIPER/internal/codegen/meta"
	"github.com/Alia5/VIIPER/internal/codegen/scanner"
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
	WireType  string
	BaseType  string
	Writer    string
	Reader    string
	IsArray   bool
	CountName string
	FixedLen  int
}

func splitWireType(wireType string) (baseType string, countToken string, isArray bool) {
	idx := strings.Index(wireType, "*")
	if idx < 0 {
		return wireType, "", false
	}
	baseType = wireType[:idx]
	countToken = wireType[idx+1:]
	return baseType, countToken, true
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
	}{Device: device, ClassName: className}
	for _, field := range tag.Fields {
		baseType, countToken, isArray := splitWireType(field.Type)
		wf := tsWireField{
			Name:     common.ToPascalCase(field.Name),
			WireType: field.Type,
			BaseType: baseType,
			Writer:   writerFor(baseType),
			Reader:   readerFor(baseType),
			IsArray:  isArray,
		}
		if isArray {
			if n, err := strconv.Atoi(countToken); err == nil {
				wf.FixedLen = n
			} else {
				wf.CountName = common.ToPascalCase(countToken)
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
{{range .Fields}}  {{.Name}}!: {{if or (eq .BaseType "u64") (eq .BaseType "i64")}}bigint{{else}}number{{end}}{{if .IsArray}}[]{{end}};
{{end}}
  constructor(init: Partial<{{.Device}}{{.ClassName}}> = {}) {
    Object.assign(this, init);
  }
  write(writer: BinaryWriter): void {
{{range .Fields}}{{if .IsArray}}{{if gt .FixedLen 0}}
		{
			const arr = (this.{{.Name}} ?? []) as any[];
			for (let i = 0; i < {{.FixedLen}}; i++) {
				writer.{{.Writer}}((arr[i] ?? 0) as any);
			}
		}
{{else}}
		for (let i = 0; i < Number((this.{{.CountName}} as any)); i++) {
			writer.{{.Writer}}((this.{{.Name}} as any[])[i]);
		}
{{end}}{{else}}    writer.{{.Writer}}(this.{{.Name}} as any);
{{end}}{{end}}  }
  static read(reader: BinaryReader): {{.Device}}{{.ClassName}} {
{{range .Fields}}{{if .IsArray}}{{if gt .FixedLen 0}}    const {{.Name | toCamelTS}}: any[] = new Array({{.FixedLen}});
		for (let i = 0; i < {{.FixedLen}}; i++) {
			{{.Name | toCamelTS}}[i] = reader.{{.Reader}}();
		}
{{else}}    const {{.Name | toCamelTS}}: any[] = new Array(Number({{.CountName | toCamelTS}}));
		for (let i = 0; i < Number({{.CountName | toCamelTS}}); i++) {
			{{.Name | toCamelTS}}[i] = reader.{{.Reader}}();
		}
{{end}}{{else}}    const {{.Name | toCamelTS}} = reader.{{.Reader}}();
{{end}}{{end}}
		return new {{.Device}}{{.ClassName}}({
{{range .Fields}}      {{.Name}}: {{.Name | toCamelTS}},
{{end}}    });
	}
}
`
