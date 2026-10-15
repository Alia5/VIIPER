package launcher

import (
	"errors"
	"os"

	"github.com/Alia5/VIIPER/cmd/launcher/pages"
	"github.com/Alia5/VIIPER/internal/config"
	"github.com/Alia5/VIIPER/meta"
	"github.com/Alia5/VIIPER/tui"

	tea "charm.land/bubbletea/v2"
	"github.com/alecthomas/kong"
	"golang.org/x/term"
)

var ErrNoTerminal = errors.New("TUI needs a terminal on stdout")

func Run() ([]string, error) {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return nil, ErrNoTerminal
	}
	parser, err := kong.New(
		&config.CLI{},
		kong.Name("VIIPER"),
		kong.Description(meta.Description()),
	)
	if err != nil {
		return nil, err
	}
	var line []string
	err = tui.Run(pages.Menu(parser, func(chosen ...string) tea.Cmd {
		line = chosen

		return tea.Quit
	}))

	return line, err
}
