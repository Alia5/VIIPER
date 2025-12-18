package csharp

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/Alia5/VIIPER/internal/codegen/common"
	"github.com/Alia5/VIIPER/internal/codegen/meta"
	"github.com/Alia5/VIIPER/internal/codegen/scanner"
)

func generateConstants(logger *slog.Logger, deviceDir string, deviceName string, md *meta.Metadata) error {
	logger.Debug("Generating constants", "device", deviceName)

	deviceConsts, ok := md.DevicePackages[deviceName]
	if !ok || deviceConsts == nil {
		logger.Warn("No constants or maps found for device", "device", deviceName)
		return nil
	}

	hasOutputSize := false
	if md.WireTags != nil {
		if s2cTag := md.WireTags.GetTag(deviceName, "s2c"); s2cTag != nil {
			if common.CalculateOutputSize(s2cTag) > 0 {
				hasOutputSize = true
			}
		}
	}

	if len(deviceConsts.Constants) == 0 && len(deviceConsts.Maps) == 0 && !hasOutputSize {
		logger.Warn("No constants or maps found for device", "device", deviceName)
		return nil
	}

	pascalDevice := toPascalCase(deviceName)
	outputPath := filepath.Join(deviceDir, pascalDevice+"Constants.cs")

	enumGroups := groupConstantsByPrefix(deviceConsts.Constants)
	maps := convertMapsToCSharp(deviceConsts.Maps)

	// Calculate OUTPUT_SIZE if device has s2c wire tag
	outputSize := 0
	if md.WireTags != nil {
		if s2cTag := md.WireTags.GetTag(deviceName, "s2c"); s2cTag != nil {
			outputSize = common.CalculateOutputSize(s2cTag)
		}
	}

	data := struct {
		Device     string
		OutputSize int
		EnumGroups []enumGroup
		Maps       []mapData
	}{
		Device:     pascalDevice,
		OutputSize: outputSize,
		Maps:       maps,
	}

	for _, eg := range enumGroups {
		if shouldGenerateEnum(eg) {
			data.EnumGroups = append(data.EnumGroups, eg)
		}
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	tmpl := template.Must(template.New("constants").Parse(constantsTemplate))
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	logger.Info("Generated constants", "device", deviceName, "path", outputPath)
	return nil
}

type enumGroup struct {
	Name      string         // Enum name (e.g., "ModifierKeys", "Buttons")
	IsFlags   bool           // Whether to use [Flags] attribute
	Type      string         // Underlying type (byte, ushort, uint)
	Constants []constantInfo // Enum members
}

type constantInfo struct {
	Name  string
	Value string
	Type  string
}

type mapData struct {
	Name      string
	KeyType   string
	ValueType string
	Entries   []mapEntry
}

type mapEntry struct {
	Key   string
	Value string
}

func groupConstantsByPrefix(constants []scanner.ConstantInfo) []enumGroup {
	groups := make(map[string]*enumGroup)

	for _, c := range constants {
		prefix := common.ExtractPrefix(c.Name)
		if prefix == "" {
			continue
		}

		if _, exists := groups[prefix]; !exists {
			groups[prefix] = &enumGroup{
				Name:      prefix,
				Constants: []constantInfo{},
			}
		}

		_, name := common.TrimPrefixAndSanitize(c.Name)
		groups[prefix].Constants = append(groups[prefix].Constants, constantInfo{
			Name:  name,
			Value: formatConstValue(c.Value),
			Type:  mapGoConstTypeToCSharp(c.Type),
		})
	}

	result := make([]enumGroup, 0, len(groups))
	for _, g := range groups {
		g.Type = inferEnumType(g.Constants)
		g.IsFlags = isFlags(g.Constants)
		result = append(result, *g)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

func shouldGenerateEnum(eg enumGroup) bool {
	return len(eg.Constants) >= 3
}

func inferEnumType(constants []constantInfo) string {
	return "uint"
}

func isFlags(constants []constantInfo) bool {
	for _, c := range constants {
		if strings.HasPrefix(c.Value, "0x") {
			var val uint64
			fmt.Sscanf(c.Value, "0x%x", &val)
			if val > 0 && (val&(val-1)) == 0 {
				return true
			}
			if val > 0 {
				return true
			}
		}
	}
	return false
}

func formatConstValue(value interface{}) string {
	switch v := value.(type) {
	case int64:
		return fmt.Sprintf("0x%X", v)
	case uint64:
		return fmt.Sprintf("0x%X", v)
	case int:
		return fmt.Sprintf("0x%X", v)
	case string:
		return fmt.Sprintf("\"%s\"", v)
	case float64:
		return fmt.Sprintf("%f", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func mapGoConstTypeToCSharp(goType string) string {
	if strings.Contains(goType, ".") {
		parts := strings.Split(goType, ".")
		return parts[len(parts)-1]
	}

	switch goType {
	case "int", "int8", "int16", "int32":
		return "int"
	case "uint8", "byte":
		return "byte"
	case "uint16":
		return "ushort"
	case "uint32", "uint":
		return "uint"
	case "uint64":
		return "ulong"
	case "string":
		return "string"
	case "char":
		return "char"
	case "bool":
		return "bool"
	default:
		return "int"
	}
}

func inferValueTypeFromEntries(entries map[string]interface{}) string {
	if len(entries) == 0 {
		return ""
	}

	allBools := true
	for _, v := range entries {
		str, ok := v.(string)
		if !ok || (str != "true" && str != "false") {
			allBools = false
			break
		}
	}
	if allBools {
		return ""
	}

	var commonPrefix string
	firstValue := true

	for _, v := range entries {
		str, ok := v.(string)
		if !ok || strings.Contains(str, " ") {
			return ""
		}

		prefix := common.ExtractPrefix(str)
		if prefix == "" {
			return ""
		}

		if firstValue {
			commonPrefix = prefix
			firstValue = false
		} else if prefix != commonPrefix {
			return ""
		}
	}

	return commonPrefix
}

func convertMapsToCSharp(maps []scanner.MapInfo) []mapData {
	result := make([]mapData, 0, len(maps))

	for _, m := range maps {
		csKeyType := mapGoConstTypeToCSharp(m.KeyType)
		csValueType := mapGoConstTypeToCSharp(m.ValueType)

		inferredValueType := inferValueTypeFromEntries(m.Entries)
		if inferredValueType != "" {
			csValueType = inferredValueType
		}

		md := mapData{
			Name:      m.Name,
			KeyType:   csKeyType,
			ValueType: csValueType,
			Entries:   make([]mapEntry, 0, len(m.Entries)),
		}

		keys := common.SortedStringKeys(m.Entries)

		for _, k := range keys {
			v := m.Entries[k]

			keyStr := formatMapKey(k, m.KeyType)
			valueStr := formatMapValue(v, m.ValueType)

			md.Entries = append(md.Entries, mapEntry{
				Key:   keyStr,
				Value: valueStr,
			})
		}

		result = append(result, md)
	}

	return result
}

func formatMapKey(key string, goType string) string {
	switch goType {
	case "byte", "uint8":
		if len(key) == 2 && key[0] == '\\' {
			switch key[1] {
			case 'n':
				return "(byte)'\\n'"
			case 'r':
				return "(byte)'\\r'"
			case 't':
				return "(byte)'\\t'"
			case '\\':
				return "(byte)'\\\\'"
			case '\'':
				return "(byte)'\\''"
			}
		}
		if len(key) == 1 {
			if key[0] >= 32 && key[0] <= 126 {
				if key[0] == '\'' {
					return "(byte)'\\''"
				} else if key[0] == '\\' {
					return "(byte)'\\\\'"
				}
				return fmt.Sprintf("(byte)'%s'", key)
			}
			return fmt.Sprintf("(byte)0x%02X", key[0])
		}
		if len(key) > 0 && (key[0] >= 'A' && key[0] <= 'Z') {
			if pfx := common.ExtractPrefix(key); pfx != "" {
				_, member := common.TrimPrefixAndSanitize(key)
				return fmt.Sprintf("(byte)%s.%s", pfx, member)
			}
		}
		return key
	case "string":
		return fmt.Sprintf("\"%s\"", key)
	default:
		return key
	}
}

func formatMapValue(value interface{}, goType string) string {
	switch goType {
	case "byte", "uint8":
		if str, ok := value.(string); ok && !strings.Contains(str, " ") {
			if pfx := common.ExtractPrefix(str); pfx != "" {
				_, member := common.TrimPrefixAndSanitize(str)
				return fmt.Sprintf("%s.%s", pfx, member)
			}
			return str
		}
		return formatConstValue(value)
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
		return formatConstValue(value)
	default:
		return formatConstValue(value)
	}
}

const constantsTemplate = `using System;
using System.Collections.Generic;

namespace Viiper.Client.Devices.{{.Device}};

{{if gt .OutputSize 0}}
/// <summary>
/// Size in bytes of {{.Device}} output (server-to-client) messages.
/// Use this constant to allocate buffers for reading device output.
/// </summary>
public static class {{.Device}}
{
    public const int OutputSize = {{.OutputSize}};
}

{{end}}
{{range .EnumGroups}}
/// <summary>
/// {{.Name}} constants for {{$.Device}} device.
/// </summary>
{{if .IsFlags}}[Flags]
{{end}}public enum {{.Name}} : {{.Type}}
{
{{range .Constants}}    {{.Name}} = {{.Value}},
{{end}}}

{{end}}
{{range .Maps}}
/// <summary>
/// {{.Name}} lookup map for {{$.Device}} device.
/// </summary>
public static class {{.Name}}
{
    private static readonly Dictionary<{{.KeyType}}, {{.ValueType}}> _map = new()
    {
{{range .Entries}}        { {{.Key}}, {{.Value}} },
{{end}}    };

    /// <summary>
    /// Try to get the value for the given key.
    /// </summary>
    public static bool TryGetValue({{.KeyType}} key, out {{.ValueType}} value)
    {
        return _map.TryGetValue(key, out value);
    }

    /// <summary>
    /// Get the value for the given key, or return the default value if not found.
    /// </summary>
    public static {{.ValueType}} GetValueOrDefault({{.KeyType}} key, {{.ValueType}} defaultValue = default)
    {
        return _map.TryGetValue(key, out var value) ? value : defaultValue;
    }

    /// <summary>
    /// Check if the map contains the given key.
    /// </summary>
    public static bool ContainsKey({{.KeyType}} key)
    {
        return _map.ContainsKey(key);
    }
}

{{end}}
`
