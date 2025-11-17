package cmd

import (
	"log/slog"
	"viiper/internal/codegen/generator"
)

type Codegen struct {
	Output string `help:"Output directory for generated SDKs (repo-root relative). Default resolves to <repo>/clients" default:"../clients" env:"VIIPER_CODEGEN_OUTPUT"`
	Lang   string `help:"Target language: c, csharp, typescript, or 'all'" default:"all" enum:"c,csharp,typescript,all" env:"VIIPER_CODEGEN_LANG"`
}

// Run is called by Kong when the codegen command is executed.
func (c *Codegen) Run(logger *slog.Logger) error {
	logger.Info("Starting VIIPER code generation", "output", c.Output, "lang", c.Lang)

	gen := generator.New(c.Output, logger)

	switch c.Lang {
	case "c":
		return gen.GenerateC()
	case "csharp":
		return gen.GenerateCSharp()
	case "typescript":
		return gen.GenerateTypeScript()
	case "all":
		if err := gen.GenerateC(); err != nil {
			return err
		}
		if err := gen.GenerateCSharp(); err != nil {
			return err
		}
		return gen.GenerateTypeScript()
	}

	return nil
}
