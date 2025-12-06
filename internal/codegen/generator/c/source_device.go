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

const deviceSourceTmpl = `/* Auto-generated VIIPER - C Client Library: device source ({{.Device}}) */

#include "viiper.h"
#include "viiper_{{.Device}}.h"

/* ========================================================================
 * {{.Device}} Map Implementations
 * ======================================================================== */
{{- range .Pkg.Maps }}

{{mapFuncImpl $.Device .}}
{{- end}}
`

type deviceSourceData struct {
	Device string
	HasS2C bool
	Pkg    *scanner.DeviceConstants
}

func generateDeviceSource(logger *slog.Logger, srcDir, device string, md *meta.Metadata) error {
	pkg := md.DevicePackages[device]
	data := deviceSourceData{
		Device: device,
		HasS2C: common.HasWireTag(md, device, "s2c"),
		Pkg:    pkg,
	}
	out := filepath.Join(srcDir, fmt.Sprintf("viiper_%s.c", device))
	t := template.Must(template.New("device.c").Funcs(tplFuncs(md)).Parse(deviceSourceTmpl))
	f, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("create device source: %w", err)
	}
	defer f.Close()
	if err := t.Execute(f, data); err != nil {
		return fmt.Errorf("exec device source tmpl: %w", err)
	}
	logger.Info("Generated device source", "device", device, "file", out)
	return nil
}
