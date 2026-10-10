package main

import (
	"os"
	"slices"
)

func init() {
	if len(os.Args) < 2 || os.Args[1] != "tui" {
		return
	}
	os.Args = slices.Delete(os.Args, 1, 2)
	launch()
}
