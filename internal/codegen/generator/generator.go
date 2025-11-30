package generator

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	cgen "github.com/Alia5/VIIPER/internal/codegen/generator/c"
	"github.com/Alia5/VIIPER/internal/codegen/generator/csharp"
	"github.com/Alia5/VIIPER/internal/codegen/generator/rust"
	"github.com/Alia5/VIIPER/internal/codegen/generator/typescript"
	"github.com/Alia5/VIIPER/internal/codegen/meta"
	"github.com/Alia5/VIIPER/internal/codegen/scanner"
)

// Generator orchestrates SDK generation for all target languages
type Generator struct {
	outputDir string
	logger    *slog.Logger
}

// LanguageGenerator defines a function that generates an SDK for a language into outputDir.
type LanguageGenerator func(logger *slog.Logger, outputDir string, md *meta.Metadata) error

var generators = map[string]LanguageGenerator{
	"c":          cgen.Generate,
	"csharp":     csharp.Generate,
	"rust":       rust.Generate,
	"typescript": typescript.Generate,
}

func New(outputDir string, logger *slog.Logger) *Generator {
	return &Generator{
		outputDir: outputDir,
		logger:    logger,
	}
}

func (g *Generator) GenAll() error {
	for lang := range generators {
		if err := g.GenerateLang(lang); err != nil {
			return fmt.Errorf("generate %s SDK: %w", lang, err)
		}
	}
	return nil
}

// GenerateLang runs the SDK generator for the provided language key.
// It scans metadata once per invocation and passes it to the language-specific generator.
func (g *Generator) GenerateLang(lang string) error {
	gen, ok := generators[lang]
	if !ok {
		var supported []string
		for k := range generators {
			supported = append(supported, k)
		}
		return fmt.Errorf("unsupported language '%s' (supported: %v)", lang, supported)
	}

	g.logger.Info("Generating SDK", "language", lang)

	md, err := g.ScanAll()
	if err != nil {
		return err
	}

	outputPath := filepath.Join(g.outputDir, lang)
	if err := os.MkdirAll(outputPath, 0o755); err != nil {
		return fmt.Errorf("failed to create %s output directory: %w", lang, err)
	}

	if err := gen(g.logger, outputPath, md); err != nil {
		return err
	}

	g.logger.Info("SDK generation complete", "language", lang, "output", outputPath)
	return nil
}

// ScanAll runs all scanners to collect metadata
func (g *Generator) ScanAll() (*meta.Metadata, error) {
	requiredPaths := []string{"internal/cmd", "apitypes", "device"}
	for _, path := range requiredPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, fmt.Errorf("codegen requires VIIPER source code and must be run from the viiper module directory: missing '%s'", path)
		}
	}

	g.logger.Info("Scanning codebase for metadata")

	md := &meta.Metadata{
		DevicePackages: make(map[string]*scanner.DeviceConstants),
		CTypeNames:     make(map[string]string),
	}

	g.logger.Debug("Scanning API routes")
	routes, err := scanner.ScanRoutesInPackage("internal/cmd")
	if err != nil {
		return nil, fmt.Errorf("failed to scan routes: %w", err)
	}
	md.Routes = routes
	g.logger.Info("Found API routes", "count", len(routes))

	g.logger.Debug("Scanning DTOs")
	dtos, err := scanner.ScanDTOsInPackage("apitypes")
	if err != nil {
		return nil, fmt.Errorf("failed to scan DTOs: %w", err)
	}
	md.DTOs = dtos
	g.logger.Info("Found DTOs", "count", len(dtos))

	// Build C type name mappings for types that would conflict
	for _, dto := range dtos {
		// Device conflicts with viiper_device_t (streaming handle), so rename to device_info
		if dto.Name == "Device" {
			md.CTypeNames["Device"] = "device_info"
		}
	}

	g.logger.Debug("Discovering device packages")
	deviceBaseDir := "device"
	entries, err := os.ReadDir(deviceBaseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read device directory: %w", err)
	}

	var devicePaths []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		deviceName := entry.Name()
		devicePath := filepath.Join(deviceBaseDir, deviceName)
		devicePaths = append(devicePaths, devicePath)

		g.logger.Debug("Scanning device package", "device", deviceName)
		deviceConsts, err := scanner.ScanDeviceConstants(devicePath)
		if err != nil {
			g.logger.Warn("Failed to scan device package", "device", deviceName, "error", err)
			continue
		}

		md.DevicePackages[deviceName] = deviceConsts
		g.logger.Info("Scanned device package",
			"device", deviceName,
			"constants", len(deviceConsts.Constants),
			"maps", len(deviceConsts.Maps))
	}

	g.logger.Debug("Scanning viiper:wire tags")
	wireTags, err := scanner.ScanWireTags(devicePaths)
	if err != nil {
		return nil, fmt.Errorf("failed to scan wire tags: %w", err)
	}
	md.WireTags = wireTags
	g.logger.Info("Scanned wire tags", "devices", len(wireTags.Tags))

	g.logger.Debug("Enriching routes with handler arg info")
	enriched, err := scanner.EnrichRoutesWithHandlerInfo(md.Routes, "internal/server/api/handler")
	if err != nil {
		return nil, fmt.Errorf("failed to enrich routes: %w", err)
	}
	md.Routes = enriched

	return md, nil
}
