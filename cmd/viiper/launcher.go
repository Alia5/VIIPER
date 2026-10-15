package main

import (
	"fmt"
	"os"
	"slices"

	"github.com/Alia5/VIIPER/cmd/launcher"
)

func launch() {
	line, err := launcher.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if line == nil {
		os.Exit(0)
	}
	os.Args = slices.Insert(os.Args, 1, line...)
}
