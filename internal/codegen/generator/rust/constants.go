package rust

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Alia5/VIIPER/internal/codegen/common"
	"github.com/Alia5/VIIPER/internal/codegen/meta"
)

const constantsTemplate = `{{.Header}}
{{if .HasMaps}}use std::collections::{{"{"}}{{if .HasHashMap}}HashMap{{end}}{{if and .HasHashMap .HasHashSet}}, {{end}}{{if .HasHashSet}}HashSet{{end}}{{"}"}};

{{end}}{{range .Constants}}pub const {{toScreamingSnakeCase .Name}}: {{.RustType}} = {{.Value}};
{{end}}
{{range .Maps}}{{if eq .ValueType "bool"}}lazy_static::lazy_static! {
    pub static ref {{toScreamingSnakeCase .Name}}: HashSet<{{.KeyRustType}}> = {
        let mut s = HashSet::new();
{{range .Entries}}        s.insert({{.Key}});
{{end}}        s
    };
}
{{else}}lazy_static::lazy_static! {
    pub static ref {{toScreamingSnakeCase .Name}}: HashMap<{{.KeyRustType}}, {{.ValueRustType}}> = {
        let mut m = HashMap::new();
{{range .Entries}}        m.insert({{.Key}}, {{.Value}});
{{end}}        m
    };
}
{{end}}
{{end}}`

type rustConstant struct {
	Name     string
	RustType string
	Value    string
}

type rustMapEntry struct {
	Key   string
	Value string
}

type rustMapInfo struct {
	Name          string
	KeyRustType   string
	ValueRustType string
	ValueType     string
	Entries       []rustMapEntry
}

type constantsData struct {
	Header     string
	DeviceName string
	Constants  []rustConstant
	Maps       []rustMapInfo
	HasMaps    bool
	HasHashMap bool
	HasHashSet bool
}

func generateConstants(logger *slog.Logger, deviceDir string, deviceName string, md *meta.Metadata) error {
	logger.Debug("Generating Rust device constants", "device", deviceName)

	devicePkg, ok := md.DevicePackages[deviceName]
	if !ok {
		return nil
	}

	var constants []rustConstant

	if md.WireTags != nil {
		if s2cTag := md.WireTags.GetTag(deviceName, "s2c"); s2cTag != nil {
			outputSize := common.CalculateOutputSize(s2cTag)
			if outputSize > 0 {
				constants = append(constants, rustConstant{
					Name:     "OUTPUT_SIZE",
					RustType: "usize",
					Value:    fmt.Sprintf("%d", outputSize),
				})
			}
		}
	}

	for _, c := range devicePkg.Constants {
		rustType := goTypeToRust(c.Type)
		value := formatConstValue(c.Value, c.Type)
		constants = append(constants, rustConstant{
			Name:     c.Name,
			RustType: rustType,
			Value:    value,
		})
	}

	var maps []rustMapInfo
	for _, m := range devicePkg.Maps {
		mapInfo := rustMapInfo{
			Name:          m.Name,
			KeyRustType:   goTypeToRust(m.KeyType),
			ValueRustType: goTypeToRust(m.ValueType),
			ValueType:     m.ValueType,
		}

		keys := common.SortedStringKeys(m.Entries)
		for _, k := range keys {
			v := m.Entries[k]
			key := formatMapKeyRust(k, m.KeyType)
			value := formatMapValueRust(v, m.ValueType)
			mapInfo.Entries = append(mapInfo.Entries, rustMapEntry{Key: key, Value: value})
		}

		if len(mapInfo.Entries) > 0 {
			maps = append(maps, mapInfo)
		}
	}

	hasHashMap := false
	hasHashSet := false
	for _, m := range maps {
		if m.ValueType == "bool" {
			hasHashSet = true
		} else {
			hasHashMap = true
		}
	}

	data := constantsData{
		Header:     writeFileHeaderRust(),
		DeviceName: deviceName,
		Constants:  constants,
		Maps:       maps,
		HasMaps:    len(maps) > 0,
		HasHashMap: hasHashMap,
		HasHashSet: hasHashSet,
	}

	funcMap := template.FuncMap{
		"toScreamingSnakeCase": toScreamingSnakeCase,
	}
	tmpl, err := template.New("constants").Funcs(funcMap).Parse(constantsTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	outputPath := filepath.Join(deviceDir, "constants.rs")
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	logger.Info("Generated device constants", "file", outputPath)
	return nil
}

func formatConstValue(val interface{}, goType string) string {
	base, _, _ := common.NormalizeGoType(goType)

	switch base {
	case "string":
		return fmt.Sprintf(`"%v"`, val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func formatMapKeyRust(key string, goType string) string {
	switch goType {
	case "byte", "uint8":
		if len(key) == 2 && key[0] == '\\' {
			switch key[1] {
			case 'n':
				return "0x0A"
			case 'r':
				return "0x0D"
			case 't':
				return "0x09"
			case '\\':
				return "0x5C"
			case '\'':
				return "0x27"
			}
		}
		if len(key) >= 1 {
			return fmt.Sprintf("%d", key[0])
		}
		return key
	case "string":
		return fmt.Sprintf("\"%s\".to_string()", key)
	default:
		return key
	}
}

func formatMapValueRust(value interface{}, goType string) string {
	switch goType {
	case "byte", "uint8":
		if str, ok := value.(string); ok {
			return toScreamingSnakeCase(str)
		}
		return fmt.Sprintf("%v", value)
	case "bool":
		if b, ok := value.(bool); ok {
			return fmt.Sprintf("%t", b)
		}
		if str, ok := value.(string); ok {
			return str
		}
		return "false"
	case "string":
		if str, ok := value.(string); ok {
			return fmt.Sprintf("\"%s\".to_string()", str)
		}
		return fmt.Sprintf("\"%v\".to_string()", value)
	default:
		return fmt.Sprintf("%v", value)
	}
}
