package typescript

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Alia5/VIIPER/internal/codegen/common"
	"github.com/Alia5/VIIPER/internal/codegen/meta"
)

func Generate(logger *slog.Logger, outputDir string, md *meta.Metadata) error {
	projectDir := outputDir
	srcDir := filepath.Join(projectDir, "src")
	typesDir := filepath.Join(srcDir, "types")
	devicesDir := filepath.Join(srcDir, "devices")
	utilsDir := filepath.Join(srcDir, "utils")

	for _, dir := range []string{projectDir, srcDir, typesDir, devicesDir, utilsDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	version, err := common.GetVersion()
	if err != nil {
		return fmt.Errorf("get version: %w", err)
	}

	if err := generateProject(logger, projectDir, version); err != nil {
		return err
	}
	if err := generateIndex(logger, srcDir); err != nil {
		return err
	}
	if err := generateBinaryUtils(logger, utilsDir); err != nil {
		return err
	}
	if err := generateAuthUtils(logger, utilsDir); err != nil {
		return err
	}
	if err := generateTypes(logger, typesDir, md); err != nil {
		return err
	}
	if err := generateClient(logger, srcDir, md); err != nil {
		return err
	}
	if err := generateDeviceWrapper(logger, srcDir); err != nil {
		return err
	}

	for deviceName := range md.DevicePackages {
		deviceDir := filepath.Join(devicesDir, common.ToPascalCase(deviceName))
		if err := os.MkdirAll(deviceDir, 0o755); err != nil {
			return fmt.Errorf("create device directory %s: %w", deviceDir, err)
		}
		if err := generateDeviceTypes(logger, deviceDir, deviceName, md); err != nil {
			return err
		}
		if err := generateConstants(logger, deviceDir, deviceName, md); err != nil {
			return err
		}
		if err := generateDeviceIndex(logger, deviceDir, deviceName); err != nil {
			return err
		}
	}

	if err := common.GenerateLicense(logger, projectDir); err != nil {
		return err
	}

	if err := common.GenerateReadme(logger, projectDir); err != nil {
		return err
	}

	logger.Info("Generated TypeScript client library", "dir", projectDir)
	return nil
}
