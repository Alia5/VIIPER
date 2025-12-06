package typescript

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const packageJSONTemplate = `{
  "name": "viiperclient",
  "version": "{{.Version}}",
  "description": "VIIPER Client Library for TypeScript/Node.js",
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "git+https://github.com/Alia5/VIIPER.git"
  },
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "files": ["dist"],
  "exports": {
    ".": {
      "types": "./dist/index.d.ts",
      "default": "./dist/index.js"
    },
    "./devices/*": {
      "types": "./dist/devices/*/index.d.ts",
      "default": "./dist/devices/*/index.js"
    }
  },
  "scripts": {
    "build": "tsc -p .",
    "clean": "rimraf dist"
  },
  "publishConfig": {
    "access": "public"
  },
  "devDependencies": {
    "@types/node": "24.10.1",
    "rimraf": "6.1.0",
    "typescript": "5.9.3"
  }
}
`

const tsconfigTemplate = `{
  "compilerOptions": {
    "target": "ES2020",
    "module": "CommonJS",
    "declaration": true,
    "outDir": "dist",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true
  },
  "include": ["src/**/*"]
}
`

func generateProject(logger *slog.Logger, projectDir, version string) error {
	logger.Debug("Generating TypeScript project scaffolding")

	if err := os.WriteFile(filepath.Join(projectDir, "package.json"), []byte(packageTemplateFor(version)), 0o644); err != nil {
		return fmt.Errorf("write package.json: %w", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "tsconfig.json"), []byte(tsconfigTemplate), 0o644); err != nil {
		return fmt.Errorf("write tsconfig.json: %w", err)
	}

	logger.Info("Generated TypeScript package.json and tsconfig.json", "version", version)
	return nil
}

func packageTemplateFor(version string) string {
	tmpl, err := template.New("pkg").Parse(packageJSONTemplate)
	if err != nil {
		return "{}"
	}
	var buf strings.Builder
	_ = tmpl.Execute(&buf, struct{ Version string }{Version: version})
	return buf.String()
}
