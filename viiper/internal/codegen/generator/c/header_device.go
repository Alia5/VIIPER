package cgen

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"
	"viiper/internal/codegen/meta"
	"viiper/internal/codegen/scanner"
)

const deviceHeaderTmpl = `#ifndef VIIPER_{{upper .Device}}_H
#define VIIPER_{{upper .Device}}_H

/* Auto-generated VIIPER - C SDK: {{.Device}} device */

#include "viiper.h"

/* ========================================================================
 * {{.Device}} Constants
 * ======================================================================== */
{{- if gt (len .Pkg.Constants) 0 }}
/* {{.Device}} constants */
{{range .Pkg.Constants -}}
#define VIIPER_{{upper $.Device}}_{{upper .Name}} {{printf "0x%X" .Value}}
{{end}}
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
	Device string
	Pkg    *scanner.DeviceConstants
}

func generateDeviceHeader(logger *slog.Logger, includeDir, device string, md *meta.Metadata) error {
	pkg := md.DevicePackages[device]
	data := deviceHeaderData{Device: device, Pkg: pkg}
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
