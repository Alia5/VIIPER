package pages

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Alia5/VIIPER/tui"

	"github.com/alecthomas/kong"
)

type launchOption struct {
	value   *kong.Value
	parsed  string
	text    string
	checked bool
}

func commandValues(parser *kong.Kong, command string) (map[string]*kong.Value, error) {
	commandCtx, err := kong.Trace(parser, strings.Fields(command))
	if err != nil {
		return nil, err
	}
	if commandCtx.Error != nil {
		return nil, commandCtx.Error
	}
	err = commandCtx.Reset()
	if err != nil {
		return nil, err
	}

	values := map[string]*kong.Value{}
	for _, flag := range commandCtx.Flags() {
		values[flag.Name] = flag.Value
	}
	for _, positional := range commandCtx.Selected().Positional {
		values[positional.Name] = positional
	}

	return values, nil
}

func valueToOption(value *kong.Value) *launchOption {
	o := &launchOption{
		value:  value,
		parsed: fmt.Sprint(value.Target.Interface()),
	}
	if value.IsBool() {
		o.checked = o.parsed == strconv.FormatBool(true)
	} else {
		o.text = o.parsed
	}

	return o
}

func optionsToElements(options []*launchOption) []tui.Element {
	var elements []tui.Element
	for _, o := range options {
		switch {
		case o.value.IsBool():
			elements = append(elements, tui.Checkbox{
				Label:    label(o.value),
				Checked:  &o.checked,
				HelpText: tui.HelpText(o.value.Help),
			})
		case o.value.Enum != "":
			elements = append(elements, &tui.Select{
				Label:    label(o.value),
				Options:  o.value.EnumSlice(),
				Selected: &o.text,
				HelpText: tui.HelpText(o.value.Help),
			})
		default:
			elements = append(elements, tui.NewText(label(o.value), o.value.Help, &o.text))
		}
	}

	return elements
}

// label turns "upstream-addr" into "Upstream Addr" and "log.level" into "Log Level".
func label(value *kong.Value) string {
	words := strings.FieldsFunc(value.Name, func(r rune) bool {
		return r == '-' || r == '.'
	})
	for i, word := range words {
		words[i] = strings.ToUpper(word[:1]) + word[1:]
	}

	return strings.Join(words, " ")
}

// commandLine passes a positional always and a flag only when it differs from its default.
func commandLine(command string, options []*launchOption) []string {
	line := strings.Fields(command)
	for _, o := range options {
		value := o.text
		if o.value.IsBool() {
			value = strconv.FormatBool(o.checked)
		}
		switch {
		case o.value.Flag == nil:
			line = append(line, value)
		case value != o.parsed:
			line = append(line, "--"+o.value.Name+"="+value)
		}
	}

	return line
}
