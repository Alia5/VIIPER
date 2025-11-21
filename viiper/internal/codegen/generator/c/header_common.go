package cgen

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"
	"viiper/internal/codegen/meta"
)

const commonHeaderTmpl = `#ifndef VIIPER_H
#define VIIPER_H

/* Auto-generated VIIPER - C SDK: common header */

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>
#include <stddef.h>

/* Platform-specific exports */
#if defined(_WIN32) || defined(_WIN64)
  #ifdef VIIPER_BUILD
    #define VIIPER_API __declspec(dllexport)
  #else
    #define VIIPER_API __declspec(dllimport)
  #endif
#else
  #define VIIPER_API __attribute__((visibility("default")))
#endif

/* Version information */
#define VIIPER_VERSION_MAJOR {{.Major}}
#define VIIPER_VERSION_MINOR {{.Minor}}
#define VIIPER_VERSION_PATCH {{.Patch}}

/* Forward declarations */
typedef struct viiper_client viiper_client_t;
typedef struct viiper_device viiper_device_t;

/* Error codes */
typedef enum {
    VIIPER_OK = 0,
    VIIPER_ERROR_CONNECT = -1,
    VIIPER_ERROR_INVALID_PARAM = -2,
    VIIPER_ERROR_PROTOCOL = -3,
    VIIPER_ERROR_TIMEOUT = -4,
    VIIPER_ERROR_MEMORY = -5,
    VIIPER_ERROR_NOT_FOUND = -6,
    VIIPER_ERROR_IO = -7
} viiper_error_t;

/* ========================================================================
 * Management API - DTOs
 * ======================================================================== */

/* Device info (special case to avoid conflict with viiper_device_t) */
typedef struct {
    uint32_t BusID;
    const char* DevId;
    const char* Vid;
    const char* Pid;
    const char* Type;
} viiper_device_info_t;

{{range .DTOs}}
{{if ne .Name "Device"}}
/* {{.Name}} */
typedef struct {
{{- range .Fields}}
    {{fieldDecl .}}
{{- end}}
} viiper_{{snakecase .Name}}_t;
{{end}}
{{end}}

/* Free helpers for DTOs (call to release memory allocated by the SDK) */
{{range .DTOs}}
{{if ne .Name "Device"}}
VIIPER_API void viiper_free_{{snakecase .Name}}(viiper_{{snakecase .Name}}_t* v);
{{end}}
{{end}}

/* ========================================================================
 * Client API
 * ======================================================================== */

/* Create a VIIPER client handle (no persistent connection for management API) */
VIIPER_API viiper_error_t viiper_client_create(
    const char* host,
    uint16_t port,
    viiper_client_t** out_client
);

/* Free the client handle and resources */
VIIPER_API void viiper_client_free(viiper_client_t* client);

/* Get the last error message */
VIIPER_API const char* viiper_get_error(viiper_client_t* client);

/* ========================================================================
 * Management API
 * ======================================================================== */
{{range .Routes}}
/* {{.Method}}: {{.Path}} */
VIIPER_API viiper_error_t viiper_{{snakecase .Handler}}(
  viiper_client_t* client{{ $params := pathParams .Path }}{{range $params}},
  const char* {{.}}{{end}}{{$payloadType := payloadCType .Payload}}{{if ne $payloadType ""}},
  {{$payloadType}} {{if eq .Payload.Kind "json"}}request{{else if eq .Payload.Kind "numeric"}}payload_value{{else}}payload_str{{end}}{{end}}{{if .ResponseDTO}},
  {{responseCType .ResponseDTO}}* out{{end}}
);
{{end}}

/* ========================================================================
 * Device Streaming API
 * ======================================================================== */

/* Callback type for receiving device output (raw bytes) */
typedef void (*viiper_output_cb)(const void* output, size_t output_size, void* user_data);

/* Create a device stream connection (opens stream socket to bus/busId/devId) */
VIIPER_API viiper_error_t viiper_device_create(
    viiper_client_t* client,
    uint32_t bus_id,
    const char* dev_id,
    viiper_device_t** out_device
);

/* Send raw input bytes to the device (client → device) */
VIIPER_API viiper_error_t viiper_device_send(
    viiper_device_t* device,
    const void* input,
    size_t input_size
);

/* Register callback for device output (device → client, async) */
VIIPER_API void viiper_device_on_output(
    viiper_device_t* device,
    viiper_output_cb callback,
    void* user_data
);

/* Close device stream and free resources */
VIIPER_API void viiper_device_close(viiper_device_t* device);

/* OpenStream: connect to an existing device's stream channel (device must already exist) */
VIIPER_API viiper_error_t viiper_open_stream(
  viiper_client_t* client,
  uint32_t bus_id,
  const char* dev_id,
  viiper_device_t** out_device
);

/* Convenience: AddDeviceAndConnect (create device then connect its stream) */
VIIPER_API viiper_error_t viiper_add_device_and_connect(
  viiper_client_t* client,
  uint32_t bus_id,
  const viiper_device_create_request_t* request,
  viiper_device_info_t* out_info,
  viiper_device_t** out_device
);

#ifdef __cplusplus
}
#endif

#endif /* VIIPER_H */
`

func generateCommonHeader(logger *slog.Logger, includeDir string, md *meta.Metadata, major, minor, patch int) error {
	out := filepath.Join(includeDir, "viiper.h")
	t := template.Must(template.New("viiper.h").Funcs(tplFuncs(md)).Parse(commonHeaderTmpl))
	f, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("create header: %w", err)
	}
	defer f.Close()

	// Create data with version and metadata
	data := struct {
		*meta.Metadata
		Major int
		Minor int
		Patch int
	}{
		Metadata: md,
		Major:    major,
		Minor:    minor,
		Patch:    patch,
	}

	if err := t.Execute(f, data); err != nil {
		return fmt.Errorf("exec header tmpl: %w", err)
	}
	logger.Info("Generated C header", "file", out)
	return nil
}
