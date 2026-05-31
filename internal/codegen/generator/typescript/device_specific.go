package typescript

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Alia5/VIIPER/internal/codegen/common"
	"github.com/Alia5/VIIPER/internal/codegen/meta"
	"github.com/Alia5/VIIPER/internal/codegen/scanner"
)

type tsDeviceSpecificField struct {
	Name      string
	TSName    string
	JSONName  string
	TSType    string
	Optional  bool
	MapAccess string
}

type tsDeviceSpecificStruct struct {
	Name           string
	FuncNamePrefix string
	Fields         []tsDeviceSpecificField
}

const deviceSpecificTemplateTS = `{{writeFileHeaderTS}}
// Typed deviceSpecific helpers for {{.Device}}

{{range .Structs}}
export interface {{.Name}} {
{{- range .Fields}}
  {{.TSName}}{{if .Optional}}?{{end}}: {{.TSType}};
{{- end}}
}

export const {{.FuncNamePrefix}}ToMap = (value: {{.Name}}): Record<string, unknown> => ({
{{- range .Fields}}
  {{.JSONName}}: value.{{.TSName}},
{{- end}}
});

export const {{.FuncNamePrefix}}FromMap = (data: Record<string, unknown> | undefined | null): {{.Name}} => {
  const map = data ?? {};
  return {
{{- range .Fields}}
    {{.TSName}}: map['{{.JSONName}}'] as {{.MapAccess}},
{{- end}}
  };
};

{{end}}
`

func generateDeviceSpecific(logger *slog.Logger, deviceDir string, deviceName string, md *meta.Metadata) error {
	structs := md.DeviceStructs[deviceName]
	if len(structs) == 0 {
		return nil
	}

	logger.Debug("Generating TS meta helpers", "device", deviceName)

	pascalDevice := common.ToPascalCase(deviceName)
	legacyPath := filepath.Join(deviceDir, pascalDevice+"DeviceSpecific.ts")
	if err := os.Remove(legacyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove legacy TS device-specific file: %w", err)
	}
	outputPath := filepath.Join(deviceDir, pascalDevice+"Meta.ts")

	data := struct {
		Device  string
		Structs []tsDeviceSpecificStruct
	}{
		Device:  pascalDevice,
		Structs: make([]tsDeviceSpecificStruct, 0, len(structs)),
	}

	for _, s := range structs {
		entry := tsDeviceSpecificStruct{
			Name:           s.Name,
			FuncNamePrefix: lowerFirst(s.Name),
			Fields:         make([]tsDeviceSpecificField, 0, len(s.Fields)),
		}
		for _, f := range s.Fields {
			tsType := fieldTypeToTSForDeviceSpecific(f)
			entry.Fields = append(entry.Fields, tsDeviceSpecificField{
				Name:      f.Name,
				TSName:    lowerFirst(f.Name),
				JSONName:  f.JSONName,
				TSType:    tsType,
				Optional:  f.Optional,
				MapAccess: tsMapAccessType(tsType),
			})
		}
		data.Structs = append(data.Structs, entry)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close() //nolint:errcheck

	tmpl := template.Must(template.New("deviceSpecificTS").Funcs(template.FuncMap{
		"writeFileHeaderTS": writeFileHeaderTS,
	}).Parse(deviceSpecificTemplateTS))

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	logger.Info("Generated TS meta helpers", "device", deviceName, "path", outputPath)
	return nil
}

func fieldTypeToTSForDeviceSpecific(field scanner.FieldInfo) string {
	typeStr := field.Type
	typeKind := field.TypeKind

	if typeKind == "map" || strings.HasPrefix(typeStr, "map[") {
		val, ok := parseGoMapType(typeStr)
		if ok {
			return "Record<string, " + goTypeToTS(val) + ">"
		}
		return "Record<string, unknown>"
	}

	if typeKind == "slice" || strings.HasPrefix(typeStr, "[]") {
		elem := strings.TrimPrefix(typeStr, "[]")
		return goTypeToTS(elem) + "[]"
	}

	if typeStr == "time.Time" {
		return "string"
	}

	if typeKind == "struct" {
		return common.ToPascalCase(typeStr)
	}

	return goTypeToTS(typeStr)
}

func tsMapAccessType(tsType string) string {
	if strings.HasPrefix(tsType, "Record<") {
		return "Record<string, unknown>"
	}
	return tsType
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}
