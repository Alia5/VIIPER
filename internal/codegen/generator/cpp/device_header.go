package cpp

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Alia5/VIIPER/internal/codegen/common"
	"github.com/Alia5/VIIPER/internal/codegen/meta"
	"github.com/Alia5/VIIPER/internal/codegen/scanner"
)

const deviceHeaderTemplate = `{{.Header}}
#pragma once

#include "../error.hpp"
#include <cstdint>
#include <vector>
{{- if .HasMaps}}
#include <string_view>
#include <array>
#include <algorithm>
#include <optional>
#include <unordered_map>
#include <unordered_set>
{{- end}}

namespace viiper {
namespace {{camelcase .DeviceName}} {

// ============================================================================
// Constants
// ============================================================================
{{if gt .OutputSize 0}}
constexpr std::size_t OUTPUT_SIZE = {{.OutputSize}};
{{end}}
{{- range .Constants}}
constexpr std::uint64_t {{.Name}} = {{.Value}};
{{- end}}
{{range .Maps}}
{{- if and (isByteKeyMap .KeyType) (hasCharLiteralKeys .Entries)}}
{{- if eq .ValueType "bool"}}
{{- $mapName := toScreamingSnakeCase .Name}}
inline const std::unordered_set<std::uint8_t> {{$mapName}} = {
{{- range $entry := sortedEntries .Entries}}
    {{formatKey $entry.Key true}},
{{- end}}
};
{{- else}}
{{- $mapName := toScreamingSnakeCase .Name}}
{{- $valueType := .ValueType}}
inline const std::unordered_map<std::uint8_t, {{cpptype $valueType}}> {{$mapName}} = {
{{- range $entry := sortedEntries .Entries}}
    { {{formatKey $entry.Key true}}, static_cast<{{cpptype $valueType}}>({{formatValue $entry.Value}}) },
{{- end}}
};
{{- end}}
{{- else if eq .ValueType "string"}}
{{- $mapName := toScreamingSnakeCase .Name}}
{{- $entries := sortedEntries .Entries}}
inline constexpr std::array<std::pair<std::uint64_t, std::string_view>, {{len $entries}}> {{$mapName}} = {{"{"}}{{"{"}}
{{- range $i, $e := $entries}}
    { {{formatKey $e.Key false}}, "{{$e.Value}}" }{{if not (isLast $i $entries)}},{{end}}
{{- end}}
}{{"}"}};

[[nodiscard]] inline std::optional<std::string_view> {{camelcase .Name}}(std::uint64_t key) noexcept {
    auto it = std::lower_bound({{$mapName}}.begin(), {{$mapName}}.end(), key,
        [](const auto& p, std::uint64_t k) { return p.first < k; });
    if (it != {{$mapName}}.end() && it->first == key) {
        return it->second;
    }
    return std::nullopt;
}
{{- else if isNumericMapVal .ValueType}}
{{- $mapName := .Name}}
{{- range $entry := sortedEntries .Entries}}
constexpr std::uint64_t {{$mapName}}_{{$entry.Key}} = {{formatValue $entry.Value}};
{{- end}}
{{- end}}
{{end}}
{{if .HasInput}}
{{$fields := wireFields .DeviceName "c2s"}}
// ============================================================================
// Input: Client -> Device
// ============================================================================

struct Input {
{{- range $fields}}
{{- if isArrayType .Type}}
    std::vector<{{cpptype (baseType .Type)}}> {{camelcase .Name}};
{{- else if not (isCountField $fields .Name)}}
    {{cpptype .Type}} {{camelcase .Name}} = 0;
{{- end}}
{{- end}}

    [[nodiscard]] std::vector<std::uint8_t> to_bytes() const {
        std::vector<std::uint8_t> buf;
{{- range $fields}}
{{- if isArrayType .Type}}
        buf.push_back(static_cast<std::uint8_t>({{camelcase .Name}}.size()));
        for (const auto& v : {{camelcase .Name}}) {
            buf.push_back(static_cast<std::uint8_t>(v));
        }
{{- else if not (isCountField $fields .Name)}}
{{- $bt := .Type}}
{{- if eq $bt "u8"}}
        buf.push_back({{camelcase .Name}});
{{- else if eq $bt "i8"}}
        buf.push_back(static_cast<std::uint8_t>({{camelcase .Name}}));
{{- else if or (eq $bt "u16") (eq $bt "i16")}}
        buf.push_back(static_cast<std::uint8_t>({{camelcase .Name}} & 0xFF));
        buf.push_back(static_cast<std::uint8_t>(({{camelcase .Name}} >> 8) & 0xFF));
{{- else if or (eq $bt "u32") (eq $bt "i32")}}
        buf.push_back(static_cast<std::uint8_t>({{camelcase .Name}} & 0xFF));
        buf.push_back(static_cast<std::uint8_t>(({{camelcase .Name}} >> 8) & 0xFF));
        buf.push_back(static_cast<std::uint8_t>(({{camelcase .Name}} >> 16) & 0xFF));
        buf.push_back(static_cast<std::uint8_t>(({{camelcase .Name}} >> 24) & 0xFF));
{{- end}}
{{- end}}
{{- end}}
        return buf;
    }
};
{{end}}
{{if .HasOutput}}
{{$fields := wireFields .DeviceName "s2c"}}
// ============================================================================
// Output: Device -> Client
// ============================================================================

struct Output {
{{- range $fields}}
    {{cpptype .Type}} {{camelcase .Name}} = 0;
{{- end}}

    static Result<Output> from_bytes(const std::uint8_t* data, std::size_t len) {
        Output result;
        std::size_t offset = 0;
{{- range $fields}}
{{- if eq .Type "u8"}}
        if (offset >= len) return Error("buffer too short");
        result.{{camelcase .Name}} = data[offset++];
{{- else if eq .Type "i8"}}
        if (offset >= len) return Error("buffer too short");
        result.{{camelcase .Name}} = static_cast<std::int8_t>(data[offset++]);
{{- else if eq .Type "u16"}}
        if (offset + 2 > len) return Error("buffer too short");
        result.{{camelcase .Name}} = data[offset] | (static_cast<std::uint16_t>(data[offset + 1]) << 8);
        offset += 2;
{{- else if eq .Type "i16"}}
        if (offset + 2 > len) return Error("buffer too short");
        result.{{camelcase .Name}} = static_cast<std::int16_t>(data[offset] | (static_cast<std::uint16_t>(data[offset + 1]) << 8));
        offset += 2;
{{- else if eq .Type "u32"}}
        if (offset + 4 > len) return Error("buffer too short");
        result.{{camelcase .Name}} = data[offset] | (static_cast<std::uint32_t>(data[offset + 1]) << 8) |
                                     (static_cast<std::uint32_t>(data[offset + 2]) << 16) | (static_cast<std::uint32_t>(data[offset + 3]) << 24);
        offset += 4;
{{- else if eq .Type "i32"}}
        if (offset + 4 > len) return Error("buffer too short");
        result.{{camelcase .Name}} = static_cast<std::int32_t>(data[offset] | (static_cast<std::uint32_t>(data[offset + 1]) << 8) |
                                     (static_cast<std::uint32_t>(data[offset + 2]) << 16) | (static_cast<std::uint32_t>(data[offset + 3]) << 24));
        offset += 4;
{{- end}}
{{- end}}
        (void)offset; // suppress unused warning
        return result;
    }
};
{{end}}

} // namespace {{camelcase .DeviceName}}
} // namespace viiper
`

func generateDeviceHeader(logger *slog.Logger, devicesDir, deviceName string, md *meta.Metadata) error {
	logger.Debug("Generating device header", "device", deviceName)
	outputFile := filepath.Join(devicesDir, deviceName+".hpp")

	devicePkg, ok := md.DevicePackages[deviceName]
	if !ok {
		return fmt.Errorf("device package %s not found in metadata", deviceName)
	}

	hasInput := md.WireTags != nil && md.WireTags.HasDirection(deviceName, "c2s")
	hasOutput := md.WireTags != nil && md.WireTags.HasDirection(deviceName, "s2c")

	funcs := tplFuncs(md)
	funcs["isLast"] = func(i int, entries []common.MapEntry) bool {
		return i == len(entries)-1
	}

	tmpl := template.Must(template.New("device").Funcs(funcs).Parse(deviceHeaderTemplate))

	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create device header: %w", err)
	}
	defer f.Close()

	hasMaps := false
	for _, m := range devicePkg.Maps {
		// Include byte-key maps (like CharToKey, ShiftChars) and string-value maps
		if isByteKeyMap(m.KeyType) || m.ValueType == "string" {
			hasMaps = true
			break
		}
	}

	// Calculate OUTPUT_SIZE from s2c wire tag
	outputSize := 0
	if md.WireTags != nil {
		if s2cTag := md.WireTags.GetTag(deviceName, "s2c"); s2cTag != nil {
			outputSize = common.CalculateOutputSize(s2cTag)
		}
	}

	data := struct {
		Header     string
		DeviceName string
		Constants  []scanner.ConstantInfo
		Maps       []scanner.MapInfo
		HasInput   bool
		HasOutput  bool
		HasMaps    bool
		OutputSize int
	}{
		Header:     writeFileHeader(),
		DeviceName: deviceName,
		Constants:  devicePkg.Constants,
		Maps:       devicePkg.Maps,
		HasInput:   hasInput,
		HasOutput:  hasOutput,
		HasMaps:    hasMaps,
		OutputSize: outputSize,
	}

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute device template: %w", err)
	}

	logger.Info("Generated device header", "device", deviceName, "file", outputFile)
	return nil
}
