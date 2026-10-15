package install

import (
	"path/filepath"

	"github.com/Alia5/VIIPER/tui"

	tea "charm.land/bubbletea/v2"
	"github.com/alecthomas/kong"
)

type wizard struct {
	target    Target
	uninstall bool
	flags     map[string]*kong.Flag
	steps     []step
	current   int
	licenses  string
	dir       string
	err       error
}

func TUI(t Target, uninstall bool, flags map[string]*kong.Flag) error {
	w := &wizard{
		target:    t,
		uninstall: uninstall,
		flags:     flags,
		steps:     installSteps,
		err:       ErrCanceled,
	}
	if uninstall {
		w.steps = uninstallSteps
	} else {
		w.dir = filepath.Dir(t.Exe)
		licenses, err := bundle.ReadFile("bundle/licenses.txt")
		if err != nil {
			return err
		}
		w.licenses = string(licenses)
	}
	err := tui.Run(w.welcome())
	if err != nil {
		return err
	}

	return w.err
}

func (w *wizard) start() tea.Cmd {
	if !w.uninstall {
		w.target.Exe = filepath.Join(w.dir, exeName)
	}

	return w.advance()
}

func (w *wizard) advance() tea.Cmd {
	for w.current < len(w.steps) && !w.steps[w.current].shouldRun(w.flags) {
		w.current++
	}
	if w.current == len(w.steps) {
		w.err = nil

		return w.finish(finishedTitle, finishedText)
	}
	title := installingTitle
	text := installingText
	if w.uninstall {
		title = uninstallingTitle
		text = uninstallingText
	}

	return tea.Sequence(tui.Show(w.progress(title, text)), w.runStep)
}

func (w *wizard) finish(title, text string) tea.Cmd {
	return tui.Show(w.progress(title, text, tui.Button{
		Label: "Finish",
		OnPressed: func() tea.Cmd {
			return tea.Quit
		},
	}))
}

func (w *wizard) runStep() tea.Msg {
	s := &w.steps[w.current]
	s.err = s.run(&w.target)
	if s.err != nil && !s.optional {
		w.err = s.err

		return w.finish(failedTitle, failedText)()
	}
	w.current++

	return w.advance()()
}
