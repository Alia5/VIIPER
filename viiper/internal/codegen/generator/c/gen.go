package cgen

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"viiper/internal/codegen/common"
	"viiper/internal/codegen/meta"
)

// Generate produces the C SDK layout under outputDir.
// It creates:
// - include/viiper/viiper.h (common types and management API)
// - include/viiper/viiper_<device>.h (per-device constants and API)
// - src/viiper.c (common implementations)
// - src/viiper_<device>.c (per-device)
// - CMakeLists.txt
func Generate(logger *slog.Logger, outputDir string, md *meta.Metadata) error {
	includeDir := filepath.Join(outputDir, "include", "viiper")
	srcDir := filepath.Join(outputDir, "src")

	if err := os.MkdirAll(includeDir, 0755); err != nil {
		return fmt.Errorf("create include dir: %w", err)
	}
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return fmt.Errorf("create src dir: %w", err)
	}

	if err := generateCommonHeader(logger, includeDir, md); err != nil {
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

	logger.Info("Generated C SDK", "dir", outputDir)
	return nil
}
