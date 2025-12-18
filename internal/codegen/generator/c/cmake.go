package cgen

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"text/template"

	"github.com/Alia5/VIIPER/internal/codegen/meta"
)

var cmakeTmpl = template.Must(template.New("cmake").Parse(`cmake_minimum_required(VERSION 3.10)
project(viiper C)

set(CMAKE_C_STANDARD 99)

# Library source files
set(VIIPER_SOURCES
    src/viiper.c
{{range .Devices}}    src/viiper_{{.}}.c
{{end}})

add_library(viiper SHARED ${VIIPER_SOURCES})
add_library(viiper_static STATIC ${VIIPER_SOURCES})

if(NOT WIN32)
    set_target_properties(viiper_static PROPERTIES OUTPUT_NAME viiper)
endif()

# Include directories
target_include_directories(viiper PUBLIC
    ${CMAKE_CURRENT_SOURCE_DIR}/include
)
target_include_directories(viiper_static PUBLIC
    ${CMAKE_CURRENT_SOURCE_DIR}/include
)

# Platform-specific settings
if(WIN32)
    target_compile_definitions(viiper PRIVATE VIIPER_BUILD)
    target_compile_definitions(viiper_static PRIVATE VIIPER_BUILD)
    target_link_libraries(viiper ws2_32)
    target_link_libraries(viiper_static ws2_32)
else()
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -fvisibility=hidden")
endif()

# Installation
install(TARGETS viiper viiper_static
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
