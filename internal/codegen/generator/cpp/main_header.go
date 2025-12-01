package cpp

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Alia5/VIIPER/internal/codegen/meta"
)

const mainHeaderTemplate = `{{.Header}}
#pragma once

// ============================================================================
// VIIPER C++ SDK - Header-Only Library
// ============================================================================
//
// Version: {{.Major}}.{{.Minor}}.{{.Patch}}
//
// Before including this file, define your JSON library configuration:
//   #define VIIPER_JSON_INCLUDE <your/json/library.hpp>
//   #define VIIPER_JSON_NAMESPACE your_namespace
//   #define VIIPER_JSON_TYPE your_json_type
//
// See config.hpp for JSON library requirements.
//
// Example with nlohmann::json:
//   #define VIIPER_JSON_INCLUDE <nlohmann/json.hpp>
//   #define VIIPER_JSON_NAMESPACE nlohmann
//   #define VIIPER_JSON_TYPE json
//   #include <viiper/viiper.hpp>
//
// ============================================================================

#include "config.hpp"
#include "error.hpp"
#include "types.hpp"
#include "client.hpp"
#include "device.hpp"

// Device-specific headers
{{range .Devices}}
#include "devices/{{.}}.hpp"
{{- end}}

namespace viiper {

// Version information
constexpr int VERSION_MAJOR = {{.Major}};
constexpr int VERSION_MINOR = {{.Minor}};
constexpr int VERSION_PATCH = {{.Patch}};

inline std::string version() {
    return std::to_string(VERSION_MAJOR) + "." +
           std::to_string(VERSION_MINOR) + "." +
           std::to_string(VERSION_PATCH);
}

} // namespace viiper
`

func generateMainHeader(logger *slog.Logger, includeDir string, md *meta.Metadata, major, minor, patch int) error {
	logger.Debug("Generating viiper.hpp")
	outputFile := filepath.Join(includeDir, "viiper.hpp")

	tmpl := template.Must(template.New("main").Funcs(tplFuncs(md)).Parse(mainHeaderTemplate))

	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create viiper.hpp: %w", err)
	}
	defer f.Close()

	devices := make([]string, 0, len(md.DevicePackages))
	for deviceName := range md.DevicePackages {
		devices = append(devices, deviceName)
	}

	data := struct {
		Header  string
		Major   int
		Minor   int
		Patch   int
		Devices []string
	}{
		Header:  writeFileHeader(),
		Major:   major,
		Minor:   minor,
		Patch:   patch,
		Devices: devices,
	}

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute main header template: %w", err)
	}

	logger.Info("Generated viiper.hpp", "file", outputFile)
	return nil
}
