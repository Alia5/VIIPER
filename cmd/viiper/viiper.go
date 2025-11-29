package main

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/Alia5/VIIPER/internal/config"
	"github.com/Alia5/VIIPER/internal/configpaths"
	"github.com/Alia5/VIIPER/internal/log"

	_ "github.com/Alia5/VIIPER/internal/registry" // Register all device handlers

	"github.com/alecthomas/kong"
	kongtoml "github.com/alecthomas/kong-toml"
	kongyaml "github.com/alecthomas/kong-yaml"
	"golang.org/x/term"
)

func main() {
	handlePlainHelpFlag()

	userCfg := findUserConfig(os.Args[1:])
	jsonPaths, yamlPaths, tomlPaths := configpaths.ConfigCandidatePaths(userCfg)

	var cli config.CLI
	ctx := kong.Parse(&cli,
		kong.Name("VIIPER"),
		kong.Description(Description()),
		kong.UsageOnError(),
		kong.Help(helpWithAsciiArt),
		// Load configuration from JSON/YAML/TOML in priority order; flags/env override config values.
		kong.Configuration(kong.JSON, jsonPaths...),
		kong.Configuration(kongyaml.Loader, yamlPaths...),
		kong.Configuration(kongtoml.Loader, tomlPaths...),
	)

	logger, closeFiles, err := log.SetupLogger(cli.Log.Level, cli.Log.File)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to setup logger:", err)
		os.Exit(2)
	}
	defer func() {
		for _, c := range closeFiles {
			_ = c.Close()
		}
	}()

	rawLogger := setupRawLogger(&cli, logger, &closeFiles)

	ctx.Bind(logger)
	ctx.BindTo(rawLogger, (*log.RawLogger)(nil))

	err = ctx.Run()
	ctx.FatalIfErrorf(err)
}

func handlePlainHelpFlag() {
	for i, arg := range os.Args[1:] {
		if arg == "-p" {
			os.Setenv("VIIPER_HELP_STYLE", "plain")
			os.Args[i+1] = "-h"
			return
		}
	}
}

func findUserConfig(args []string) string {
	for i, a := range args {
		if strings.HasPrefix(a, "--config=") {
			return a[len("--config="):]
		}
		if a == "--config" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return os.Getenv("VIIPER_CONFIG")
}

func setupRawLogger(cli *config.CLI, logger *slog.Logger, closeFiles *[]io.Closer) log.RawLogger {
	if cli.Log.RawFile != "" {
		f, err := os.OpenFile(cli.Log.RawFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			logger.Error("failed to open raw log file", "file", cli.Log.RawFile, "error", err)
			return log.NewRaw(nil)
		}
		*closeFiles = append(*closeFiles, f)
		return log.NewRaw(f)
	}
	if cli.Log.Level == "trace" {
		return log.NewRaw(os.Stdout)
	}
	return log.NewRaw(nil)
}

func helpWithAsciiArt(options kong.HelpOptions, ctx *kong.Context) error {
	// VIIPER_HELP_STYLE env var: "plain", "big", "small", or auto-detect
	helpStyle := strings.ToLower(os.Getenv("VIIPER_HELP_STYLE"))
	if helpStyle == "" {
		helpStyle = detectHelpStyle()
	}
	if helpStyle == "plain" {
		return kong.DefaultHelpPrinter(options, ctx)
	}

	helpText := captureHelpOutput(options, ctx)

	art := asciiBrailleColoredSmall
	if helpStyle == "big" {
		art = asciiBrailleColoredBig
	}

	output := mergeArtWithHelp(normalizeLineEndings(art), normalizeLineEndings(helpText))
	_, err := fmt.Fprint(ctx.Stdout, output)
	return err
}

func captureHelpOutput(options kong.HelpOptions, ctx *kong.Context) string {
	var buf bytes.Buffer
	origStdout := ctx.Stdout
	ctx.Stdout = &buf
	_ = kong.DefaultHelpPrinter(options, ctx)
	ctx.Stdout = origStdout
	return buf.String()
}

func normalizeLineEndings(s string) string {
	s = strings.TrimRight(s, "\r\n")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\r", "\n")
}

func mergeArtWithHelp(art, help string) string {
	artLines := strings.Split(art, "\n")
	helpLines := strings.Split(help, "\n")

	artWidth := maxVisibleWidth(artLines) + 2

	maxLines := max(len(artLines), len(helpLines))
	artOffset := (len(helpLines) - len(artLines)) / 2
	if artOffset < 0 {
		artOffset = 0
	}

	var out strings.Builder
	for i := range maxLines {
		artLine := ""
		if idx := i - artOffset; idx >= 0 && idx < len(artLines) {
			artLine = artLines[idx]
		}

		helpLine := ""
		if i < len(helpLines) {
			helpLine = helpLines[i]
		}

		padding := artWidth - visibleWidth(artLine)
		out.WriteString(artLine)
		out.WriteString(strings.Repeat(" ", padding))
		out.WriteString(helpLine)
		out.WriteString("\n")
	}
	return out.String()
}

func maxVisibleWidth(lines []string) int {
	maxWidth := 0
	for _, line := range lines {
		if w := visibleWidth(line); w > maxWidth {
			maxWidth = w
		}
	}
	return maxWidth
}

func visibleWidth(s string) int {
	inEscape := false
	width := 0
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		width++
	}
	return width
}

func detectHelpStyle() string {
	fd := int(os.Stdout.Fd())
	if !term.IsTerminal(fd) {
		fd = int(os.Stderr.Fd())
		if !term.IsTerminal(fd) {
			return "plain"
		}
	}

	if os.Getenv("TERM") == "dumb" {
		return "plain"
	}

	width, _, err := term.GetSize(fd)
	if err != nil || width <= 0 {
		return "small"
	}

	const (
		bigThreshold   = 140
		smallThreshold = 110
	)
	switch {
	case width >= bigThreshold:
		return "big"
	case width >= smallThreshold:
		return "small"
	default:
		return "plain"
	}
}
