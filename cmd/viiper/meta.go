package main

import (
	_ "embed"
	"fmt"
	"runtime/debug"
	"time"
)

//go:embed ascii_braille_colored_sm.txt
var asciiBrailleColoredSmall string

//go:embed ascii_braille_colored.txt
var asciiBrailleColoredBig string

var (
	Version = ""
	Commit  = ""
	Date    = ""
)

var descriptionTemplate = `
Virtual Input over IP EmulatoR
  Version: %s (%s)
           %s
  Source:  https://github.com/Alia5/VIIPER
  License: GPLv3
`

func Description() string {
	return fmt.Sprintf(descriptionTemplate, Version, Commit, Date)
}

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		if Version == "" {
			Version = info.Main.Version
			if Version == "" || Version == "(devel)" {
				Version = "dev"
			}
		}
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if Commit == "" {
					if len(setting.Value) > 7 {
						Commit = setting.Value[:7]
					} else {
						Commit = setting.Value
					}
				}
			case "vcs.time":
				if Date == "" {
					if t, err := time.Parse(time.RFC3339, setting.Value); err == nil {
						Date = t.Format("2006-01-02")
					} else {
						Date = setting.Value
					}
				}
			}
		}
	}
	if Version == "" {
		Version = "dev"
	}
	if Commit == "" {
		Commit = "unknown"
	}
	if Date == "" {
		Date = "unknown"
	}
}
