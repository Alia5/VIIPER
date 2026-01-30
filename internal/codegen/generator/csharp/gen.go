package csharp

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Alia5/VIIPER/internal/codegen/common"
	"github.com/Alia5/VIIPER/internal/codegen/meta"
)

func Generate(logger *slog.Logger, outputDir string, md *meta.Metadata) error {
	projectDir := filepath.Join(outputDir, "Viiper.Client")
	typesDir := filepath.Join(projectDir, "Types")
	devicesDir := filepath.Join(projectDir, "Devices")
	examplesDir := filepath.Join(outputDir, "examples")

	for _, dir := range []string{projectDir, typesDir, devicesDir, examplesDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	version, err := common.GetVersion()
	if err != nil {
		return fmt.Errorf("get version: %w", err)
	}

	if err := generateProject(logger, projectDir, md, version); err != nil {
		return err
	}

	if err := generateTypes(logger, typesDir, md); err != nil {
		return err
	}

	if err := generateClient(logger, projectDir, md); err != nil {
		return err
	}
	if err := generateAuthHelper(logger, projectDir); err != nil {
		return err
	}
	if err := generateDevice(logger, projectDir, md); err != nil {
		return err
	}

	for deviceName := range md.DevicePackages {
		deviceDir := filepath.Join(devicesDir, toPascalCase(deviceName))
		if err := os.MkdirAll(deviceDir, 0755); err != nil {
			return fmt.Errorf("create device directory %s: %w", deviceDir, err)
		}

		if err := generateDeviceTypes(logger, deviceDir, deviceName, md); err != nil {
			return err
		}

		if err := generateConstants(logger, deviceDir, deviceName, md); err != nil {
			return err
		}
	}

	if err := common.GenerateLicense(logger, outputDir); err != nil {
		return err
	}

	if err := common.GenerateReadme(logger, outputDir); err != nil {
		return err
	}

	logger.Info("Generated C# client library", "dir", outputDir)
	return nil
}
