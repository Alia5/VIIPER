package common

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

const readmeTemplate = `# VIIPER Client Library

This is an automatically generated client library for [VIIPER](https://github.com/Alia5/VIIPER) - **Virtual** **I**nput over **IP** **E**mulato**R**

## Documentation

- **Project Repository**: https://github.com/Alia5/VIIPER
- **Documentation**: https://alia5.github.io/VIIPER/

## About VIIPER

VIIPER creates virtual USB input devices using the USBIP protocol.  
These virtual devices appear as real hardware to the operating system and applications, allowing you to emulate controllers, keyboards, and other input devices without physical hardware.

## License

MIT License - See LICENSE.txt for details.
`

func GenerateReadme(logger *slog.Logger, outputDir string) error {
	readmePath := filepath.Join(outputDir, "README.md")

	if err := os.WriteFile(readmePath, []byte(readmeTemplate), 0644); err != nil {
		return fmt.Errorf("write README.md: %w", err)
	}

	logger.Debug("Generated README.md", "path", readmePath)
	return nil
}
