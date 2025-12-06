package cpp

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

const configTemplate = `// Auto-generated VIIPER C++ Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase

#pragma once

// ============================================================================
// JSON Parser Configuration
// ============================================================================
//
// VIIPER C++ Client Library requires a JSON library for parsing API responses.
// You must define VIIPER_JSON_INCLUDE before including viiper.hpp
//
// Option 1: Use nlohmann::json (recommended)
//   #define VIIPER_JSON_INCLUDE <nlohmann/json.hpp>
//   #define VIIPER_JSON_NAMESPACE nlohmann
//   #define VIIPER_JSON_TYPE json
//   #include <viiper/viiper.hpp>
//
// Option 2: Use custom JSON library
//   Define VIIPER_JSON_INCLUDE, VIIPER_JSON_NAMESPACE, and VIIPER_JSON_TYPE
//   Your JSON type must support:
//   - parse(const std::string&) -> JsonType
//   - dump() -> std::string
//   - operator[](const std::string&) -> JsonType
//   - contains(const std::string&) -> bool
//   - is_number(), is_string(), is_array(), is_object() -> bool
//   - get<T>() -> T
//   - size() -> std::size_t (for arrays)
//
// ============================================================================

#ifndef VIIPER_JSON_INCLUDE
#error "VIIPER_JSON_INCLUDE must be defined before including viiper.hpp (e.g., #define VIIPER_JSON_INCLUDE <nlohmann/json.hpp>)"
#endif

#ifndef VIIPER_JSON_NAMESPACE
#error "VIIPER_JSON_NAMESPACE must be defined before including viiper.hpp (e.g., #define VIIPER_JSON_NAMESPACE nlohmann)"
#endif

#ifndef VIIPER_JSON_TYPE
#error "VIIPER_JSON_TYPE must be defined before including viiper.hpp (e.g., #define VIIPER_JSON_TYPE json)"
#endif

#include VIIPER_JSON_INCLUDE

namespace viiper {
namespace json_ns = VIIPER_JSON_NAMESPACE;
using json_type = json_ns::VIIPER_JSON_TYPE;
} // namespace viiper
`

func generateConfig(logger *slog.Logger, includeDir string) error {
	logger.Debug("Generating config.hpp")
	outputFile := filepath.Join(includeDir, "config.hpp")

	if err := os.WriteFile(outputFile, []byte(configTemplate), 0644); err != nil {
		return fmt.Errorf("write config.hpp: %w", err)
	}

	logger.Info("Generated config.hpp", "file", outputFile)
	return nil
}
