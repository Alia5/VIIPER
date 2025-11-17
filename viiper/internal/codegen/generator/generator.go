package generator

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	cgen "viiper/internal/codegen/generator/c"
	"viiper/internal/codegen/meta"
	"viiper/internal/codegen/scanner"
)

// Generator orchestrates SDK generation for all target languages
type Generator struct {
	outputDir string
	logger    *slog.Logger
}

// New creates a new Generator instance
func New(outputDir string, logger *slog.Logger) *Generator {
	return &Generator{
		outputDir: outputDir,
		logger:    logger,
	}
}

// ScanAll runs all scanners to collect metadata
func (g *Generator) ScanAll() (*meta.Metadata, error) {
	g.logger.Info("Scanning codebase for metadata")

	md := &meta.Metadata{
		DevicePackages: make(map[string]*scanner.DeviceConstants),
	}

	// Scan routes
	g.logger.Debug("Scanning API routes")
	routes, err := scanner.ScanRoutesInPackage("internal/cmd")
	if err != nil {
		return nil, fmt.Errorf("failed to scan routes: %w", err)
	}
	md.Routes = routes
	g.logger.Info("Found API routes", "count", len(routes))

	// Scan DTOs
	g.logger.Debug("Scanning DTOs")
	dtos, err := scanner.ScanDTOsInPackage("pkg/apitypes")
	if err != nil {
		return nil, fmt.Errorf("failed to scan DTOs: %w", err)
	}
	md.DTOs = dtos
	g.logger.Info("Found DTOs", "count", len(dtos))

	// Discover and scan device packages
	g.logger.Debug("Discovering device packages")
	deviceBaseDir := "pkg/device"
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

	// Scan wire tags from all device packages
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

// GenerateC generates the C SDK
func (g *Generator) GenerateC() error {
	g.logger.Info("Generating C SDK")

	md, err := g.ScanAll()
	if err != nil {
		return err
	}

	outputPath := filepath.Join(g.outputDir, "c")
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("failed to create C output directory: %w", err)
	}

	// Delegate to C generator subpackage
	if err := cgen.Generate(g.logger, outputPath, md); err != nil {
		return err
	}

	g.logger.Info("C SDK generation complete", "output", outputPath)
	return nil
}

// GenerateCSharp generates the C# SDK
func (g *Generator) GenerateCSharp() error {
	g.logger.Info("Generating C# SDK")
	// TODO: Implement in Step 12
	return fmt.Errorf("C# generation not yet implemented")
}

// GenerateTypeScript generates the TypeScript SDK
func (g *Generator) GenerateTypeScript() error {
	g.logger.Info("Generating TypeScript SDK")
	// TODO: Implement in Step 13
	return fmt.Errorf("TypeScript generation not yet implemented")
}
