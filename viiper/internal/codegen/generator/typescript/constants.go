package typescript

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"viiper/internal/codegen/common"
	"viiper/internal/codegen/meta"
	"viiper/internal/codegen/scanner"
)

type tsEnumGroup struct {
	Name      string
	Type      string
	Constants []tsConstInfo
}

type tsConstInfo struct {
	Name  string
	Value string
}

type tsMapData struct {
	Name      string
	KeyType   string
	ValueType string
	Entries   []tsMapEntry
}

type tsMapEntry struct {
	Key   string
	Value string
}

func generateConstants(logger *slog.Logger, deviceDir string, deviceName string, md *meta.Metadata) error {
	deviceConsts, ok := md.DevicePackages[deviceName]
	if !ok || deviceConsts == nil {
		return nil
	}
	if len(deviceConsts.Constants) == 0 && len(deviceConsts.Maps) == 0 {
		return nil
	}
	pascalDevice := common.ToPascalCase(deviceName)
	outputPath := filepath.Join(deviceDir, pascalDevice+"Constants.ts")
	enums := groupConstants(deviceConsts.Constants)
	maps := convertMaps(deviceConsts.Maps)
	data := struct {
		Device string
		Enums  []tsEnumGroup
		Maps   []tsMapData
	}{Device: pascalDevice, Enums: enums, Maps: maps}
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()
	tmpl := template.Must(template.New("constsTS").Funcs(template.FuncMap{
		"writeFileHeaderTS": writeFileHeaderTS,
	}).Parse(constantsTemplateTS))
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}
	logger.Info("Generated TS constants", "device", deviceName, "path", outputPath)
	return nil
}

func groupConstants(constants []scanner.ConstantInfo) []tsEnumGroup {
	groups := map[string]*tsEnumGroup{}
	for _, c := range constants {
		prefix := common.ExtractPrefix(c.Name)
		if prefix == "" {
			continue
		}
		g := groups[prefix]
		if g == nil {
			g = &tsEnumGroup{Name: prefix, Type: "number"}
			groups[prefix] = g
		}
		name := strings.TrimPrefix(c.Name, prefix)
		if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
			name = "Num" + name
		}
		g.Constants = append(g.Constants, tsConstInfo{Name: name, Value: formatConstValueTS(c.Value)})
	}
	var result []tsEnumGroup
	for _, g := range groups {
		if len(g.Constants) >= 3 {
			result = append(result, *g)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

func formatConstValueTS(v interface{}) string {
	switch t := v.(type) {
	case int64:
		return fmt.Sprintf("0x%X", t)
	case uint64:
		return fmt.Sprintf("0x%X", t)
	case int:
		return fmt.Sprintf("0x%X", t)
	case string:
		return fmt.Sprintf("'%s'", t)
	case float64:
		return fmt.Sprintf("%f", t)
	default:
		return fmt.Sprintf("%v", t)
	}
}

func convertMaps(maps []scanner.MapInfo) []tsMapData {
	var result []tsMapData
	for _, m := range maps {
		md := tsMapData{Name: m.Name, KeyType: mapGoConstTypeToTS(m.KeyType), ValueType: mapGoConstTypeToTS(m.ValueType)}
		keys := make([]string, 0, len(m.Entries))
		for k := range m.Entries {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := m.Entries[k]
			md.Entries = append(md.Entries, tsMapEntry{Key: formatMapKeyTS(k, m.KeyType), Value: formatMapValueTS(v, m.ValueType)})
		}
		result = append(result, md)
	}
	return result
}

func mapGoConstTypeToTS(goType string) string {
	if strings.Contains(goType, ".") {
		return goType[strings.LastIndex(goType, ".")+1:]
	}
	switch goType {
	case "int", "int8", "int16", "int32", "uint8", "byte", "uint16", "uint32", "uint64":
		return "number"
	case "string":
		return "string"
	case "bool":
		return "boolean"
	default:
		return "number"
	}
}

func formatMapKeyTS(key string, goType string) string {
	switch goType {
	case "byte", "uint8":
		// Prefer enum symbolic key if provided (e.g., Key.A)
		if len(key) > 0 && (key[0] >= 'A' && key[0] <= 'Z') {
			prefix := common.ExtractPrefix(key)
			if prefix != "" {
				member := strings.TrimPrefix(key, prefix)
				if len(member) > 0 {
					if member[0] >= '0' && member[0] <= '9' {
						member = "Num" + member
					}
					return fmt.Sprintf("%s.%s", prefix, member)
				}
			}
		}
		// Otherwise, emit numeric key code for literal chars/escapes
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
			return fmt.Sprintf("0x%02X", key[0])
		}
		return key
	case "string":
		return fmt.Sprintf("\"%s\"", key)
	default:
		return key
	}
}

func formatMapValueTS(value interface{}, goType string) string {
	switch goType {
	case "byte", "uint8":
		if str, ok := value.(string); ok && !strings.Contains(str, " ") {
			prefix := common.ExtractPrefix(str)
			if prefix != "" {
				member := strings.TrimPrefix(str, prefix)
				if len(member) > 0 && member[0] >= '0' && member[0] <= '9' {
					member = "Num" + member
				}
				return fmt.Sprintf("%s.%s", prefix, member)
			}
			return str
		}
		return formatConstValueTS(value)
	case "bool":
		if b, ok := value.(bool); ok {
			if b {
				return "true"
			}
			return "false"
		}
		if str, ok := value.(string); ok {
			return str
		}
		return "false"
	case "string":
		if str, ok := value.(string); ok {
			return fmt.Sprintf("\"%s\"", str)
		}
		return formatConstValueTS(value)
	default:
		return formatConstValueTS(value)
	}
}

const constantsTemplateTS = `{{writeFileHeaderTS}}
// {{.Device}} constants and maps

{{range .Enums}}
export enum {{.Name}} {
{{range .Constants}}  {{.Name}} = {{.Value}},
{{end}}}
{{end}}

{{range .Maps}}
export const {{.Name}}: Record<number, {{.ValueType}}> = {
{{range .Entries}}  [{{.Key}}]: {{.Value}},
{{end}}};

export const {{.Name}}Get = (key: number, def?: {{.ValueType}}): {{.ValueType}} | undefined => {
	return Object.prototype.hasOwnProperty.call({{.Name}}, key) ? {{.Name}}[key] : def;
};

export const {{.Name}}Has = (key: number): boolean => Object.prototype.hasOwnProperty.call({{.Name}}, key);
{{end}}
`
