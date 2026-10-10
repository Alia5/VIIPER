package install

import (
	"fmt"
	"strings"

	"github.com/Alia5/VIIPER/internal/tui"
)

const (
	installingTitle   = "Installing"
	installingText    = "Please wait while Setup installs VIIPER on your computer."
	uninstallingTitle = "Uninstalling"
	uninstallingText  = "Please wait while VIIPER is removed from your computer."
	finishedTitle     = "Completing the VIIPER Setup Wizard"
	finishedText      = "Setup has finished. Select Finish to exit Setup."
	failedTitle       = "Setup Failed"
	failedText        = "Setup could not complete. See the step marked [!] below."
)

func (w *wizard) progress(title, text string, rows ...tui.Row) tui.Page {
	return tui.Page{
		Title: title,
		Text:  text,
		Hint:  "Enter: confirm   Esc: quit",
		Rows: append([]tui.Row{
			{
				View: func(int) string {
					var b strings.Builder
					for i, s := range w.steps {
						switch {
						case s.optional && !s.enabled:
							fmt.Fprintln(&b, "[-] "+s.label)
						case s.err != nil:
							fmt.Fprintf(&b, "[!] %s: %v\n", s.label, s.err)
						case i < w.current:
							fmt.Fprintln(&b, "[x] "+s.label)
						case i == w.current:
							fmt.Fprintf(&b, "%s %s\n", w.spin.View(), s.label)
						default:
							fmt.Fprintln(&b, "[ ] "+s.label)
						}
					}

					return strings.TrimSuffix(b.String(), "\n")
				},
			},
		}, rows...),
	}
}
