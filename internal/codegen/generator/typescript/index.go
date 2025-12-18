package typescript

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Alia5/VIIPER/internal/codegen/common"
)

const indexTemplate = `{{writeFileHeaderTS}}
export * from './ViiperClient';
export * from './ViiperDevice';
export * as Types from './types/ManagementDtos';
export * as Keyboard from './devices/Keyboard';
export * as Mouse from './devices/Mouse';
export * as Xbox360 from './devices/Xbox360';
`

const deviceIndexTemplate = `{{writeFileHeaderTS}}
export * from './{{.PascalName}}Input';
{{if .HasOutput}}export * from './{{.PascalName}}Output';
{{end}}export * from './{{.PascalName}}Constants';
`

func generateIndex(logger *slog.Logger, srcDir string) error {
	logger.Debug("Generating index.ts re-exports")
	f, err := os.Create(filepath.Join(srcDir, "index.ts"))
	if err != nil {
		return fmt.Errorf("write index.ts: %w", err)
	}
	defer f.Close()
	tmpl := template.Must(template.New("index").Funcs(template.FuncMap{
		"writeFileHeaderTS": writeFileHeaderTS,
	}).Parse(indexTemplate))
	if err := tmpl.Execute(f, nil); err != nil {
		return fmt.Errorf("execute index template: %w", err)
	}
	return nil
}

func generateDeviceIndex(logger *slog.Logger, deviceDir, deviceName string) error {
	logger.Debug("Generating device index.ts", "device", deviceName)

	pascalName := common.ToPascalCase(deviceName)

	hasOutput := false
	outputPath := filepath.Join(deviceDir, pascalName+"Output.ts")
	if _, err := os.Stat(outputPath); err == nil {
		hasOutput = true
	}

	f, err := os.Create(filepath.Join(deviceDir, "index.ts"))
	if err != nil {
		return fmt.Errorf("write device index.ts: %w", err)
	}
	defer f.Close()

	tmpl := template.Must(template.New("deviceIndex").Funcs(template.FuncMap{
		"writeFileHeaderTS": writeFileHeaderTS,
	}).Parse(deviceIndexTemplate))

	data := struct {
		PascalName string
		HasOutput  bool
	}{
		PascalName: pascalName,
		HasOutput:  hasOutput,
	}

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute device index template: %w", err)
	}
	return nil
}
