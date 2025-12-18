package cmd

import (
	"log/slog"

	"github.com/Alia5/VIIPER/internal/codegen/generator"
)

type Codegen struct {
	Output string `help:"Output directory for generated client libraries (repo-root relative). Default resolves to <repo>/clients" default:"./clients" env:"VIIPER_CODEGEN_OUTPUT"`
	Lang   string `help:"Target language: c, cpp, csharp, rust, typescript, or 'all'" default:"all" enum:"c,cpp,csharp,rust,typescript,all" env:"VIIPER_CODEGEN_LANG"`
}

// Run is called by Kong when the codegen command is executed.
func (c *Codegen) Run(logger *slog.Logger) error {
	logger.Info("Starting VIIPER code generation", "output", c.Output, "lang", c.Lang)

	gen := generator.New(c.Output, logger)
	if c.Lang == "all" {
		return gen.GenAll()
	}
	return gen.GenerateLang(c.Lang)

}
