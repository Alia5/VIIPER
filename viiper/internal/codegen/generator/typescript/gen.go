package typescript

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"viiper/internal/codegen/common"
	"viiper/internal/codegen/meta"
)

// Generate produces the TypeScript SDK layout under outputDir.
// It creates a Node package with the following structure:
// - package.json, tsconfig.json
// - src/index.ts
// - src/ViiperClient.ts (management API)
// - src/ViiperDevice.ts (device stream wrapper)
// - src/types/ManagementDtos.ts (DTOs)
// - src/devices/<Device>/{<Device>Input.ts,<Device>Output.ts,<Device>Constants.ts}
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

	logger.Info("Generated TypeScript SDK", "dir", projectDir)
	return nil
}
