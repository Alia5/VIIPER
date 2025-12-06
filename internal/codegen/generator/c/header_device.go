package cgen

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

const deviceHeaderTmpl = `#ifndef VIIPER_{{upper .Device}}_H
#define VIIPER_{{upper .Device}}_H

/* Auto-generated VIIPER - C Client Library: {{.Device}} device */

/* Minimal path fix: include common header without duplicated module path */
#include "viiper.h"

/* ========================================================================
 * {{.Device}} Constants
 * ======================================================================== */
/* Device output size (bytes to read from socket in callback) */
#define VIIPER_{{upper .Device}}_OUTPUT_SIZE {{.OutputSize}}

{{- if gt (len .Pkg.Constants) 0 }}
/* {{.Device}} constants */
{{range .Pkg.Constants -}}
#define VIIPER_{{upper $.Device}}_{{upper (snakecase .Name)}} {{printf "0x%X" .Value}}
{{end}}
{{- end}}

/* ========================================================================
 * {{.Device}} Lookup Maps
 * ======================================================================== */
{{- range .Pkg.Maps }}
/* {{.Name}} lookup function */
{{mapFuncDecl $.Device .}}
{{- end}}

/* ========================================================================
 * {{.Device}} Input/Output Structures
 * ======================================================================== */
{{if hasWireTag .Device "c2s"}}
#pragma pack(push, 1)
typedef struct {
{{wireFields .Device "c2s" | indent 4}}
} viiper_{{.Device}}_input_t;
#pragma pack(pop)
{{end}}

{{if hasWireTag .Device "s2c"}}
#pragma pack(push, 1)
typedef struct {
{{wireFields .Device "s2c" | indent 4}}
} viiper_{{.Device}}_output_t;
#pragma pack(pop)
{{end}}

#endif /* VIIPER_{{upper .Device}}_H */
`

type deviceHeaderData struct {
	Device     string
	Pkg        *scanner.DeviceConstants
	OutputSize int
}

func generateDeviceHeader(logger *slog.Logger, includeDir, device string, md *meta.Metadata) error {
	pkg := md.DevicePackages[device]

	// Calculate OUTPUT_SIZE from s2c wire tag
	outputSize := 0
	if md.WireTags != nil {
		if s2cTag := md.WireTags.GetTag(device, "s2c"); s2cTag != nil {
			outputSize = common.CalculateOutputSize(s2cTag)
		}
	}

	data := deviceHeaderData{
		Device:     device,
		Pkg:        pkg,
		OutputSize: outputSize,
	}
	out := filepath.Join(includeDir, fmt.Sprintf("viiper_%s.h", device))
	t := template.Must(template.New("device.h").Funcs(tplFuncs(md)).Parse(deviceHeaderTmpl))
	f, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("create device header: %w", err)
	}
	defer f.Close()
	if err := t.Execute(f, data); err != nil {
		return fmt.Errorf("exec device header tmpl: %w", err)
	}
	logger.Info("Generated device header", "device", device, "file", out)
	return nil
}
