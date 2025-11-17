package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"viiper/internal/codegen/scanner"
)

func main() {
	projectRoot, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	apitypesPkg := filepath.Join(projectRoot, "pkg", "apitypes")
	schemas, err := scanner.ScanDTOsInPackage(apitypesPkg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to scan DTOs: %v\n", err)
		os.Exit(1)
	}

	output, err := json.MarshalIndent(schemas, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}
