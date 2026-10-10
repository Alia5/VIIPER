package cmd

import (
	"os"
	"path/filepath"

	"github.com/Alia5/VIIPER/internal/install"
	"golang.org/x/term"
)

type Install struct {
	Quiet      bool   `short:"q" help:"No TUI, defaults for every choice"`
	InstallDir string `help:"Install directory"`
	NoUpdate   bool   `help:"Disable auto-update upon install"`
}

type Uninstall struct {
	Quiet bool `short:"q" help:"No TUI"`
}

func (c *Install) Run(t install.Target, banner string) error {
	exe, err := install.DefaultExe()
	if err != nil {
		return err
	}
	if c.InstallDir != "" {
		exe = filepath.Join(c.InstallDir, filepath.Base(exe))
	}
	t.Exe = exe
	t.NoUpdate = c.NoUpdate

	return run(c.Quiet, t, false, banner)
}

func (c *Uninstall) Run(t install.Target, banner string) error {
	t.Exe = t.Source

	return run(c.Quiet, t, true, banner)
}

func run(quiet bool, t install.Target, uninstall bool, banner string) error {
	if quiet || !term.IsTerminal(int(os.Stdout.Fd())) {
		return install.Run(t, uninstall)
	}

	return install.TUI(t, uninstall, banner)
}
