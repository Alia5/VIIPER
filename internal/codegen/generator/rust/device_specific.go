package rust

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

type rustDeviceSpecificField struct {
	JSONName string
	RustName string
	RustType string
	Optional bool
}

type rustDeviceSpecificStruct struct {
	Name   string
	Fields []rustDeviceSpecificField
}

const deviceSpecificTemplate = `{{.Header}}
use std::collections::HashMap;

{{range .Structs}}
#[derive(Debug, Clone, Default, serde::Serialize, serde::Deserialize)]
pub struct {{.Name}} {
{{- range .Fields}}
    {{if .Optional}}#[serde(skip_serializing_if = "Option::is_none")]
	{{end}}#[serde(rename = "{{.JSONName}}")]
	pub {{.RustName}}: {{.RustType}},
{{- end}}
}

impl {{.Name}} {
    pub fn to_map(&self) -> HashMap<String, serde_json::Value> {
        match serde_json::to_value(self) {
            Ok(serde_json::Value::Object(obj)) => obj.into_iter().collect(),
            _ => HashMap::new(),
        }
    }

    pub fn from_map(map: &HashMap<String, serde_json::Value>) -> Result<Self, serde_json::Error> {
        let obj: serde_json::Map<String, serde_json::Value> = map.clone().into_iter().collect();
        serde_json::from_value(serde_json::Value::Object(obj))
    }
}

{{end}}
`

func generateDeviceSpecific(logger *slog.Logger, deviceDir string, deviceName string, md *meta.Metadata) error {
	structs := md.DeviceStructs[deviceName]
	if len(structs) == 0 {
		return nil
	}

	logger.Debug("Generating Rust meta helpers", "device", deviceName)

	legacyPath := filepath.Join(deviceDir, "device_specific.rs")
	if err := os.Remove(legacyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove legacy Rust device-specific file: %w", err)
	}
	outputPath := filepath.Join(deviceDir, "meta.rs")
	data := struct {
		Header  string
		Structs []rustDeviceSpecificStruct
	}{
		Header:  writeFileHeaderRust(),
		Structs: make([]rustDeviceSpecificStruct, 0, len(structs)),
	}

	for _, s := range structs {
		entry := rustDeviceSpecificStruct{
			Name:   s.Name,
			Fields: make([]rustDeviceSpecificField, 0, len(s.Fields)),
		}
		for _, f := range s.Fields {
			rustName := common.ToSnakeCase(f.Name)
			if isRustKeyword(rustName) {
				rustName = "r#" + rustName
			}
			rustType := fieldTypeToRustForDeviceSpecific(f)
			entry.Fields = append(entry.Fields, rustDeviceSpecificField{
				JSONName: f.JSONName,
				RustName: rustName,
				RustType: rustType,
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

	tmpl := template.Must(template.New("deviceSpecificRust").Parse(deviceSpecificTemplate))
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	logger.Info("Generated Rust meta helpers", "device", deviceName, "path", outputPath)
	return nil
}

func fieldTypeToRustForDeviceSpecific(field scanner.FieldInfo) string {
	typeStr := field.Type
	typeKind := field.TypeKind
	optional := field.Optional

	var rustType string
	if typeKind == "map" || strings.HasPrefix(typeStr, "map[") {
		valueType, ok := parseGoMapType(typeStr)
		if ok {
			rustType = fmt.Sprintf("std::collections::HashMap<String, %s>", goTypeToRust(valueType))
		} else {
			rustType = "std::collections::HashMap<String, serde_json::Value>"
		}
	} else if typeKind == "slice" || strings.HasPrefix(typeStr, "[]") {
		elem := strings.TrimPrefix(typeStr, "[]")
		rustType = fmt.Sprintf("Vec<%s>", goTypeToRust(elem))
	} else if typeStr == "time.Time" {
		rustType = "String"
	} else {
		rustType = goTypeToRust(typeStr)
	}

	if optional && !strings.HasPrefix(rustType, "Option<") {
		rustType = fmt.Sprintf("Option<%s>", rustType)
	}

	return rustType
}
