package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"viiper/internal/configpaths"

	toml "github.com/pelletier/go-toml"
	yaml "gopkg.in/yaml.v3"
)

// ConfigCommand groups config-related subcommands.
type ConfigCommand struct {
	Init ConfigInit `cmd:"" help:"Generate a configuration template"`
}

// ConfigInit scaffolds a configuration file for a specific command.
type ConfigInit struct {
	Command string `arg:"" name:"command" help:"Command to generate config for" enum:"server,proxy"`
	Format  string `help:"Output format" enum:"json,yaml,toml" default:"json"`
	Output  string `help:"Destination file path (defaults to current directory)"`
	Force   bool   `help:"Overwrite if the file already exists"`
}

// Run generates a configuration template dynamically via reflection of the command structs and tags.
func (c *ConfigInit) Run() error {
	format := normalizeFormat(c.Format)
	if format == "" {
		return fmt.Errorf("unsupported format: %s", c.Format)
	}

	var root map[string]any
	switch c.Command {
	case "server":
		root = buildMapFromStruct(reflect.TypeOf(Server{}))
	case "proxy":
		root = buildMapFromStruct(reflect.TypeOf(Proxy{}))
	default:
		return errors.New("unknown command; expected 'server' or 'proxy'")
	}

	dest := c.Output
	if dest == "" {
		ext := "json"
		if format == "yaml" {
			ext = "yaml"
		} else if format == "toml" {
			ext = "toml"
		}
		dest = c.Command + "." + ext
	}

	if !c.Force {
		if _, err := os.Stat(dest); err == nil {
			return errors.New("destination exists; use --force to overwrite")
		}
	}
	if err := configpaths.EnsureDir(dest); err != nil {
		return err
	}

	var data []byte
	var err error
	switch format {
	case "json":
		data, err = json.MarshalIndent(root, "", "  ")
	case "yaml":
		data, err = yaml.Marshal(root)
	case "toml":
		data, err = toml.Marshal(root)
	}
	if err != nil {
		return err
	}
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return err
	}
	return nil
}

func normalizeFormat(f string) string {
	switch strings.ToLower(f) {
	case "json":
		return "json"
	case "yaml", "yml":
		return "yaml"
	case "toml":
		return "toml"
	default:
		return ""
	}
}

func lowerCamel(s string) string {
	if s == "" {
		return s
	}
	// Convert first character to lowercase
	r := []rune(s)
	r[0] = toLower(r[0])
	return string(r)
}

func toLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}

func buildMapFromStruct(t reflect.Type) map[string]any {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	out := map[string]any{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		if f.Tag.Get("kong") == "-" {
			continue
		}

		if _, ok := f.Tag.Lookup("embed"); ok {
			prefix := f.Tag.Get("prefix")
			name := strings.TrimSuffix(prefix, ".")
			sub := buildMapFromStruct(f.Type)
			if name != "" {
				out[name] = sub
			} else {
				for k, v := range sub {
					out[k] = v
				}
			}
			continue
		}

		key := lowerCamel(f.Name)
		def := f.Tag.Get("default")
		val := defaultValueForField(f.Type, def)
		if val != nil {
			out[key] = val
		}
	}
	return out
}

func defaultValueForField(t reflect.Type, def string) any {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.PkgPath() == "time" && t.Name() == "Duration" {
		if def != "" {
			return def
		}
		return "0s"
	}
	switch t.Kind() {
	case reflect.String:
		return def // may be empty
	case reflect.Bool:
		if def == "" {
			return false
		}
		b, err := strconv.ParseBool(def)
		if err != nil {
			return false
		}
		return b
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if def == "" {
			return 0
		}
		n, err := strconv.ParseInt(def, 10, 64)
		if err != nil {
			return 0
		}
		return n
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if def == "" {
			return 0
		}
		n, err := strconv.ParseUint(def, 10, 64)
		if err != nil {
			return 0
		}
		return n
	case reflect.Float32, reflect.Float64:
		if def == "" {
			return 0
		}
		f, err := strconv.ParseFloat(def, 64)
		if err != nil {
			return 0
		}
		return f
	case reflect.Struct:
		return buildMapFromStruct(t)
	default:
		return nil
	}
}
