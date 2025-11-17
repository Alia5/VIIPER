package cgen

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"
	"viiper/internal/codegen/meta"
)

const deviceSourceTmpl = `/* Auto-generated VIIPER - C SDK: device source ({{.Device}}) */

#include "viiper/viiper.h"
#include "viiper/viiper_{{.Device}}.h"

`

type deviceSourceData struct {
	Device string
	HasS2C bool
}

func generateDeviceSource(logger *slog.Logger, srcDir, device string, md *meta.Metadata) error {
	data := deviceSourceData{Device: device, HasS2C: hasWireTag(md, device, "s2c")}
	out := filepath.Join(srcDir, fmt.Sprintf("viiper_%s.c", device))
	t := template.Must(template.New("device.c").Parse(deviceSourceTmpl))
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
