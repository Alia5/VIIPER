package cgen

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"text/template"
	"viiper/internal/codegen/meta"
)

var cmakeTmpl = template.Must(template.New("cmake").Parse(`cmake_minimum_required(VERSION 3.10)
project(viiper C)

set(CMAKE_C_STANDARD 99)

# Library source files
add_library(viiper SHARED
    src/viiper.c
{{range .Devices}}    src/viiper_{{.}}.c
{{end}})

# Include directories
target_include_directories(viiper PUBLIC
    ${CMAKE_CURRENT_SOURCE_DIR}/include
)

# Platform-specific settings
if(WIN32)
    target_compile_definitions(viiper PRIVATE VIIPER_BUILD)
    target_link_libraries(viiper ws2_32)
else()
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -fvisibility=hidden")
endif()

# Installation
install(TARGETS viiper
    LIBRARY DESTINATION lib
    ARCHIVE DESTINATION lib
    RUNTIME DESTINATION bin
)

install(DIRECTORY include/viiper
    DESTINATION include
)
`))

func generateCMake(logger *slog.Logger, outDir string, md *meta.Metadata) error {
	devices := make([]string, 0, len(md.DevicePackages))
	for device := range md.DevicePackages {
		devices = append(devices, device)
	}
	sort.Strings(devices)

	data := struct {
		Devices []string
	}{
		Devices: devices,
	}

	var buf bytes.Buffer
	if err := cmakeTmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("execute CMake template: %w", err)
	}

	out := filepath.Join(outDir, "CMakeLists.txt")
	if err := os.WriteFile(out, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("write CMakeLists.txt: %w", err)
	}
	logger.Info("Generated CMakeLists.txt", "file", out)
	return nil
}
