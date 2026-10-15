package uninstall

import (
	"os"

	"github.com/Alia5/VIIPER/internal/install"

	"golang.org/x/term"
)

type Uninstall struct {
	Quiet bool `short:"q" help:"No TUI"`
}

func (u *Uninstall) Run(t install.Target) error {
	t.Exe = t.Source
	if u.Quiet || !term.IsTerminal(int(os.Stdout.Fd())) {
		return install.Run(t, true, nil)
	}

	return install.TUI(t, true, nil)
}
