package cpp

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

	includeDir := filepath.Join(outputDir, "include", "viiper")
	detailDir := filepath.Join(includeDir, "detail")
	devicesDir := filepath.Join(includeDir, "devices")

	for _, dir := range []string{includeDir, detailDir, devicesDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	if err := generateConfig(logger, includeDir); err != nil {
		return err
	}

	if err := generateError(logger, includeDir); err != nil {
		return err
	}

	if err := generateTypes(logger, includeDir, md); err != nil {
		return err
	}

	if err := generateSocket(logger, detailDir); err != nil {
		return err
	}

	if err := generateJson(logger, detailDir); err != nil {
		return err
	}

	if err := generateClient(logger, includeDir, md); err != nil {
		return err
	}

	if err := generateDevice(logger, includeDir, md); err != nil {
		return err
	}

	for deviceName := range md.DevicePackages {
		if err := generateDeviceHeader(logger, devicesDir, deviceName, md); err != nil {
			return err
		}
	}

	if err := generateMainHeader(logger, includeDir, md, major, minor, patch); err != nil {
		return err
	}

	if err := common.GenerateLicense(logger, outputDir); err != nil {
		return err
	}

	if err := common.GenerateReadme(logger, outputDir); err != nil {
		return err
	}

	logger.Info("Generated C++ client library", "dir", outputDir)
	return nil
}
