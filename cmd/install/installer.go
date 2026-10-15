package install

import (
	"os"
	"path/filepath"

	"github.com/Alia5/VIIPER/internal/install"

	"github.com/alecthomas/kong"
	"golang.org/x/term"
)

type Installer struct {
	Quiet      bool   `short:"q" help:"No TUI, defaults for every choice"`
	InstallDir string `help:"Install directory"`
	Update     bool   `name:"update" help:"Update to the latest release before installing" default:"true" negatable:""`
}

func (i *Installer) Run(ctx *kong.Context, t install.Target) error {
	exe, err := install.DefaultExe()
	if err != nil {
		return err
	}
	if i.InstallDir != "" {
		exe = filepath.Join(i.InstallDir, filepath.Base(exe))
	}
	t.Exe = exe
	flags := map[string]*kong.Flag{}
	for _, flag := range ctx.Flags() {
		flags[flag.Name] = flag
	}
	if i.Quiet || !term.IsTerminal(int(os.Stdout.Fd())) {
		return install.Run(t, false, flags)
	}

	return install.TUI(t, false, flags)
}
