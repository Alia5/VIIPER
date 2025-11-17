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

	serverFile := filepath.Join(projectRoot, "internal", "cmd", "server.go")
	routes, err := scanner.ScanRoutes(serverFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to scan routes: %v\n", err)
		os.Exit(1)
	}

	handlerPkg := filepath.Join(projectRoot, "internal", "server", "api", "handler")
	enrichedRoutes, err := scanner.EnrichRoutesWithHandlerInfo(routes, handlerPkg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to enrich routes: %v\n", err)
		os.Exit(1)
	}

	output, err := json.MarshalIndent(enrichedRoutes, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}
