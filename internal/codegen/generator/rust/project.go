package rust

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"
)

const cargoTomlTemplate = `[package]
name = "viiper-client"
version = "{{.Version}}"
edition = "2021"
rust-version = "1.70"
authors = ["Peter Repukat"]
description = "VIIPER Client Library for Rust"
license = "MIT"
repository = "https://github.com/Alia5/VIIPER"
homepage = "https://github.com/Alia5/VIIPER"
documentation = "https://alia5.github.io/VIIPER/stable/clients/rust/"
readme = "README.md"
keywords = ["viiper", "usbip", "virtual-device", "input-emulation", "hid"]
categories = ["api-bindings", "hardware-support"]

[dependencies]
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
thiserror = "2.0"
lazy_static = "1.5"

[dependencies.tokio]
version = "1.0"
features = ["net", "io-util", "rt", "time", "macros"]
optional = true

[dependencies.tokio-util]
version = "0.7"
optional = true

[features]
default = []
async = ["tokio", "tokio-util"]

[package.metadata.docs.rs]
all-features = true
`

func generateProject(logger *slog.Logger, projectDir string, version string) error {
	logger.Debug("Generating Cargo.toml")
	outputFile := filepath.Join(projectDir, "Cargo.toml")

	funcMap := template.FuncMap{}
	tmpl, err := template.New("cargo").Funcs(funcMap).Parse(cargoTomlTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	data := struct {
		Version string
	}{
		Version: version,
	}

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	logger.Info("Generated Cargo.toml", "file", outputFile)
	return nil
}

func generateGitignore(logger *slog.Logger, projectDir string) error {
	logger.Debug("Generating .gitignore")
	outputFile := filepath.Join(projectDir, ".gitignore")

	content := `# Rust build artifacts
target/
Cargo.lock

# IDE
.vscode/
.idea/
*.swp
*.swo
*~
`

	if err := os.WriteFile(outputFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("write .gitignore: %w", err)
	}

	logger.Info("Generated .gitignore", "file", outputFile)
	return nil
}
