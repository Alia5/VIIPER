package cgen

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Alia5/VIIPER/internal/codegen/common"
	"github.com/Alia5/VIIPER/internal/codegen/meta"
)

func Generate(logger *slog.Logger, outputDir string, md *meta.Metadata) error {
	version, err := common.GetVersion()
	if err != nil {
		return fmt.Errorf("get version: %w", err)
	}
	major, minor, patch := common.ParseVersion(version)
	logger.Info("Using version", "version", version, "major", major, "minor", minor, "patch", patch)

	includeDir := filepath.Join(outputDir, "include")
	srcDir := filepath.Join(outputDir, "src")

	if err := os.MkdirAll(includeDir, 0755); err != nil {
		return fmt.Errorf("create include dir: %w", err)
	}
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return fmt.Errorf("create src dir: %w", err)
	}

	if err := generateCommonHeader(logger, includeDir, md, major, minor, patch); err != nil {
		return err
	}

	for device := range md.DevicePackages {
		if err := generateDeviceHeader(logger, includeDir, device, md); err != nil {
			return err
		}
	}

	if err := generateCommonSource(logger, srcDir, md); err != nil {
		return err
	}

	for device := range md.DevicePackages {
		if err := generateDeviceSource(logger, srcDir, device, md); err != nil {
			return err
		}
	}

	if err := generateCMake(logger, outputDir, md); err != nil {
		return err
	}

	if err := common.GenerateLicense(logger, outputDir); err != nil {
		return err
	}

	if err := common.GenerateReadme(logger, outputDir); err != nil {
		return err
	}

	logger.Info("Generated C SDK", "dir", outputDir)
	return nil
}
