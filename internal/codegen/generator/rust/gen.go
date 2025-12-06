package rust

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
	devicesDir := filepath.Join(srcDir, "devices")

	for _, dir := range []string{projectDir, srcDir, devicesDir} {
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

	if err := generateGitignore(logger, projectDir); err != nil {
		return err
	}

	if err := generateError(logger, srcDir); err != nil {
		return err
	}

	if err := generateWireModule(logger, srcDir); err != nil {
		return err
	}

	if err := generateTypes(logger, srcDir, md); err != nil {
		return err
	}

	if err := generateClient(logger, srcDir, md); err != nil {
		return err
	}

	if err := generateAsyncClient(logger, srcDir, md); err != nil {
		return err
	}

	for deviceName := range md.DevicePackages {
		deviceDir := filepath.Join(devicesDir, deviceName)
		if err := os.MkdirAll(deviceDir, 0o755); err != nil {
			return fmt.Errorf("create device directory %s: %w", deviceDir, err)
		}

		if err := generateDeviceTypes(logger, deviceDir, deviceName, md); err != nil {
			return err
		}

		if err := generateConstants(logger, deviceDir, deviceName, md); err != nil {
			return err
		}

		if err := generateDeviceModFile(logger, deviceDir, deviceName, md); err != nil {
			return err
		}
	}

	if err := generateDevicesModFile(logger, devicesDir, md); err != nil {
		return err
	}

	if err := generateLibFile(logger, srcDir, md); err != nil {
		return err
	}

	if err := common.GenerateLicense(logger, projectDir); err != nil {
		return err
	}

	if err := common.GenerateReadme(logger, projectDir); err != nil {
		return err
	}

	logger.Info("Generated Rust client library", "dir", projectDir)
	return nil
}

func generateLibFile(logger *slog.Logger, srcDir string, md *meta.Metadata) error {
	logger.Debug("Generating lib.rs")
	outputFile := filepath.Join(srcDir, "lib.rs")

	content := writeFileHeaderRust() + `
pub mod error;
pub mod wire;
pub mod types;
pub mod client;

#[cfg(feature = "async")]
pub mod async_client;

pub mod devices;

pub use error::{ViiperError, ProblemJson};
pub use wire::{DeviceInput, DeviceOutput};
pub use types::*;
pub use client::{ViiperClient, DeviceStream};

#[cfg(feature = "async")]
pub use async_client::{AsyncViiperClient, AsyncDeviceStream};
`

	if err := os.WriteFile(outputFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("write lib.rs: %w", err)
	}

	logger.Info("Generated lib.rs", "file", outputFile)
	return nil
}

func generateDevicesModFile(logger *slog.Logger, devicesDir string, md *meta.Metadata) error {
	logger.Debug("Generating devices/mod.rs")
	outputFile := filepath.Join(devicesDir, "mod.rs")

	content := writeFileHeaderRust()
	for deviceName := range md.DevicePackages {
		content += fmt.Sprintf("pub mod %s;\n", deviceName)
	}

	if err := os.WriteFile(outputFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("write devices/mod.rs: %w", err)
	}

	logger.Info("Generated devices/mod.rs", "file", outputFile)
	return nil
}

func generateDeviceModFile(logger *slog.Logger, deviceDir string, deviceName string, md *meta.Metadata) error {
	logger.Debug("Generating device mod.rs", "device", deviceName)
	outputFile := filepath.Join(deviceDir, "mod.rs")

	content := writeFileHeaderRust()

	hasInput := md.WireTags != nil && md.WireTags.GetTag(deviceName, "c2s") != nil
	hasOutput := md.WireTags != nil && md.WireTags.GetTag(deviceName, "s2c") != nil
	hasConstants := md.DevicePackages[deviceName] != nil &&
		(len(md.DevicePackages[deviceName].Constants) > 0 || len(md.DevicePackages[deviceName].Maps) > 0)

	if hasInput {
		content += "pub mod input;\n"
		content += "pub use input::*;\n\n"
	}
	if hasOutput {
		content += "pub mod output;\n"
		content += "pub use output::*;\n\n"
	}
	if hasConstants {
		content += "pub mod constants;\n"
		content += "pub use constants::*;\n\n"
	}

	if err := os.WriteFile(outputFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("write device mod.rs: %w", err)
	}

	logger.Info("Generated device mod.rs", "device", deviceName, "file", outputFile)
	return nil
}
