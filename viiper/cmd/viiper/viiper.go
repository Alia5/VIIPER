package main

import (
	"os"
	"viiper/internal/config"
	"viiper/internal/log"

	_ "viiper/internal/registry" // Register all device handlers

	"github.com/alecthomas/kong"
)

func main() {
	var cli config.CLI
	ctx := kong.Parse(&cli,
		kong.Name("viiper"),
		kong.Description("Virtual Input over IP EmulatoR"),
		kong.UsageOnError(),
	)

	logger, closeFiles, err := log.SetupLogger(cli.Log.Level, cli.Log.File)
	if err != nil {
		_, _ = os.Stderr.WriteString("failed to setup logger: " + err.Error() + "\n")
		os.Exit(2)
	}
	defer func() {
		for _, c := range closeFiles {
			_ = c.Close()
		}
	}()

	var rawLogger log.RawLogger
	if cli.Log.RawFile != "" {
		f, err := os.OpenFile(cli.Log.RawFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			logger.Error("failed to open raw log file", "file", cli.Log.RawFile, "error", err)
			rawLogger = log.NewRaw(nil)
		} else {
			rawLogger = log.NewRaw(f)
			closeFiles = append(closeFiles, f)
		}
	} else if cli.Log.Level == "trace" {
		rawLogger = log.NewRaw(os.Stdout)
	} else {
		rawLogger = log.NewRaw(nil)
	}

	ctx.Bind(logger)
	ctx.BindTo(rawLogger, (*log.RawLogger)(nil))

	err = ctx.Run()
	ctx.FatalIfErrorf(err)
}
