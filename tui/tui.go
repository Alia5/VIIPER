package tui

import (
	"errors"
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	paddingX      = 2
	paddingY      = 1
	logoGap       = 4
	buttonGap     = "  "
	minTextHeight = 1
	minTextWidth  = 1
	foldedHeight  = 1
	focusArrow    = "➔ "
	focusBlank    = "  "
	closedArrow   = " ▼"
	openArrow     = " ▲"
	accentColor   = "#2a9db1"
)

type Element interface {
	View(focused bool) string
	Help() string
	Update(msg tea.Msg) tea.Cmd
}

type HelpText string

type Page struct {
	Title    string
	Text     string
	Hint     string
	Steps    []Step
	Elements []Element
	Buttons  []Button
}

type screen struct {
	start  tea.Cmd
	page   Page
	shown  int
	text   viewport.Model
	spin   spinner.Model
	focus  int
	width  int
	height int
	err    error
}

type box struct {
	x      int
	y      int
	width  int
	height int
}

type hit struct {
	shown int
	index int
	mouse tea.Msg
}

var (
	keyQuit     = key.NewBinding(key.WithKeys("ctrl+c"))
	keyCancel   = key.NewBinding(key.WithKeys("esc"))
	keyNext     = key.NewBinding(key.WithKeys("tab", "down"))
	keyPrevious = key.NewBinding(key.WithKeys("shift+tab", "up"))
	keyRight    = key.NewBinding(key.WithKeys("right"))
	keyLeft     = key.NewBinding(key.WithKeys("left"))
	keyConfirm  = key.NewBinding(key.WithKeys("enter", "space"))

	textKeys = viewport.KeyMap{
		PageUp:   key.NewBinding(key.WithKeys("pgup")),
		PageDown: key.NewBinding(key.WithKeys("pgdown")),
	}

	titleStyle    = lipgloss.NewStyle().Bold(true)
	focusStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(accentColor)).Bold(true)
	buttonStyle   = lipgloss.NewStyle().Padding(0, 2)
	selectedStyle = buttonStyle.Padding(0, 1).
			Background(lipgloss.Color(accentColor)).Foreground(lipgloss.BrightWhite).
			Bold(true).
			BorderStyle(
			lipgloss.Border{
				Left:  "",
				Right: "",
			}).BorderForeground(lipgloss.Color(accentColor)).
		BorderRight(true).BorderLeft(true).BorderTop(false).BorderBottom(false)
	helpStyle  = lipgloss.NewStyle().Faint(true).Italic(true)
	hintStyle  = lipgloss.NewStyle().Faint(true)
	thumbStyle = lipgloss.NewStyle().Faint(true)
	trackStyle = lipgloss.NewStyle().Faint(true)
)

func Run(start tea.Cmd) error {
	s := &screen{
		start: start,
		spin:  spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(focusStyle)),
	}
	_, err := tea.NewProgram(s).Run()

	return errors.Join(err, s.err)
}

func Show(page Page) tea.Cmd {
	return func() tea.Msg {
		return page
	}
}

func Fail(err error) tea.Cmd {
	return func() tea.Msg {
		return err
	}
}

func (h HelpText) Help() string {
	return string(h)
}

func shouldActivate(msg tea.Msg) bool {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return key.Matches(msg, keyConfirm)
	case tea.MouseClickMsg:
		return msg.Button == tea.MouseLeft
	}

	return false
}

func (b box) contains(x, y int) bool {
	return x >= b.x && x < b.x+b.width && y >= b.y && y < b.y+b.height
}

func (b box) relative(msg tea.MouseMsg) tea.Msg {
	mouse := msg.Mouse()
	mouse.X -= b.x
	mouse.Y -= b.y
	switch msg.(type) {
	case tea.MouseClickMsg:
		return tea.MouseClickMsg(mouse)
	case tea.MouseMotionMsg:
		return tea.MouseMotionMsg(mouse)
	}

	return nil
}

func (s *screen) Init() tea.Cmd {
	return s.start
}

func (s *screen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height

		return s, nil
	case Page:
		s.page = msg
		s.shown++
		s.focus = 0
		s.text = viewport.New()
		s.text.KeyMap = textKeys
		if s.shouldSpin() {
			return s, s.spin.Tick
		}

		return s, nil
	case spinner.TickMsg:
		if !s.shouldSpin() {
			return s, nil
		}
		var cmd tea.Cmd
		s.spin, cmd = s.spin.Update(msg)

		return s, cmd
	case error:
		s.err = msg

		return s, tea.Quit
	case hit:
		if msg.shown != s.shown || s.dropdownOpen() && msg.index != s.focus {
			return s, nil
		}
		s.focus = msg.index

		return s, s.focused().Update(msg.mouse)
	case tea.MouseWheelMsg:
		s.text, _ = s.text.Update(msg)

		return s, nil
	case tea.MouseMsg:
		return s, nil
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keyQuit):
			return s, tea.Quit
		case s.dropdownOpen():
			return s, s.focused().Update(msg)
		case key.Matches(msg, keyCancel):
			return s, tea.Quit
		case key.Matches(msg, textKeys.PageUp, textKeys.PageDown):
			s.text, _ = s.text.Update(msg)

			return s, nil
		case key.Matches(msg, keyNext):
			s.move(1)

			return s, nil
		case key.Matches(msg, keyPrevious):
			s.move(-1)

			return s, nil
		case key.Matches(msg, keyRight) && s.focus >= len(s.page.Elements) && s.focus+1 < s.focusable():
			s.focus++

			return s, nil
		case key.Matches(msg, keyLeft) && s.focus > len(s.page.Elements):
			s.focus--

			return s, nil
		}
	}
	focused := s.focused()
	if focused == nil {
		return s, nil
	}

	return s, focused.Update(msg)
}

func (s *screen) View() tea.View {
	content, boxes := s.render()
	shown := s.shown
	view := tea.NewView(content)
	view.AltScreen = true
	view.MouseMode = tea.MouseModeAllMotion
	view.OnMouse = func(msg tea.MouseMsg) tea.Cmd {
		mouse := msg.Mouse()
		index := slices.IndexFunc(boxes, func(b box) bool {
			return b.contains(mouse.X, mouse.Y)
		})
		if index < 0 {
			return nil
		}
		relative := boxes[index].relative(msg)
		if relative == nil {
			return nil
		}

		return func() tea.Msg {
			return hit{
				shown: shown,
				index: index,
				mouse: relative,
			}
		}
	}

	return view
}

func (s *screen) render() (string, []box) {
	var (
		head     []string
		body     []string
		boxes    []box
		unfolded int
	)
	logo := logoFor(s.width)
	room := s.width - 2*paddingX
	if logo != "" {
		room -= lipgloss.Width(logo) + logoGap
	}
	room = max(room, minTextWidth)
	if s.page.Title != "" {
		head = append(head, strings.Split(titleStyle.Render(lipgloss.Wrap(s.page.Title, room, "")), "\n")...)
		head = append(head, "")
	}
	if s.page.Text != "" {
		body = append(body, "")
	}
	for _, step := range s.page.Steps {
		body = append(body, strings.Split(lipgloss.Wrap(s.stepView(step), room, ""), "\n")...)
	}
	if len(s.page.Steps) > 0 && s.focusable() > 0 {
		body = append(body, "")
	}
	for i, e := range s.page.Elements {
		view := e.View(i == s.focus)
		unfolded += lipgloss.Height(view) - foldedHeight
		boxes = append(boxes, box{
			y:      len(body),
			width:  lipgloss.Width(view),
			height: lipgloss.Height(view),
		})
		body = append(body, strings.Split(view, "\n")...)
	}
	if len(s.page.Elements) > 0 && len(s.page.Buttons) > 0 {
		body = append(body, "")
	}
	if len(s.page.Buttons) > 0 {
		var (
			row []string
			x   int
		)
		for i, b := range s.page.Buttons {
			view := b.View(len(s.page.Elements)+i == s.focus)
			boxes = append(boxes, box{
				x:      x,
				y:      len(body),
				width:  lipgloss.Width(view),
				height: lipgloss.Height(view),
			})
			row = append(row, view)
			x += lipgloss.Width(view) + len(buttonGap)
		}
		body = append(body, strings.Join(row, buttonGap))
	}
	help := ""
	if focused := s.focused(); focused != nil {
		help = focused.Help()
	}
	body = append(body, "")
	body = append(body, strings.Split(helpStyle.Render(lipgloss.Wrap(help, room, "")), "\n")...)
	body = append(body, "")
	body = append(body, strings.Split(hintStyle.Render(lipgloss.Wrap(s.page.Hint, room, "")), "\n")...)
	if s.page.Text != "" {
		wrapped := lipgloss.Wrap(s.page.Text, max(room-lipgloss.Width(scrollbarGap+scrollbarThumb), minTextWidth), "")
		s.text.SetContent(wrapped)
		s.text.SetWidth(max(lipgloss.Width(wrapped), minTextWidth))
		s.text.SetHeight(min(s.text.TotalLineCount(), max(s.height-2*paddingY-len(head)-len(body), minTextHeight)))
		text := s.text.View()
		if s.text.TotalLineCount() > s.text.Height() {
			text = lipgloss.JoinHorizontal(lipgloss.Top, text, scrollbarGap, s.scrollbar())
		}
		head = append(head, strings.Split(text, "\n")...)
	}
	x, y := paddingX, paddingY
	framed := strings.Join(append(head, body...), "\n")
	if logo != "" {
		centred := max(min(lipgloss.Height(logo), s.height-2*paddingY)-len(head)-len(body)+unfolded, 0) / 2
		x += lipgloss.Width(logo) + logoGap
		y += centred
		framed = lipgloss.JoinHorizontal(
			lipgloss.Top,
			logo,
			lipgloss.NewStyle().PaddingLeft(logoGap).PaddingTop(centred).Render(framed),
		)
	}
	for i := range boxes {
		boxes[i].x += x
		boxes[i].y += y + len(head)
	}

	return lipgloss.NewStyle().Padding(paddingY, paddingX).Render(framed), boxes
}

func (s *screen) focusable() int {
	return len(s.page.Elements) + len(s.page.Buttons)
}

func (s *screen) focused() Element {
	switch {
	case s.focus < len(s.page.Elements):
		return s.page.Elements[s.focus]
	case s.focus < s.focusable():
		return s.page.Buttons[s.focus-len(s.page.Elements)]
	}

	return nil
}

func (s *screen) dropdownOpen() bool {
	dropdown, ok := s.focused().(*Select)

	return ok && dropdown.open
}

func (s *screen) move(step int) {
	count := s.focusable()
	if count == 0 {
		return
	}
	s.focus = (s.focus + step + count) % count
}
