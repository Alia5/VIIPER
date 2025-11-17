package common

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

const mitLicenseTemplate = `MIT License

Copyright (c) %d Peter Repukat

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`

func GenerateLicense(logger *slog.Logger, outputDir string) error {
	licensePath := filepath.Join(outputDir, "LICENSE.txt")

	currentYear := time.Now().Year()
	licenseText := fmt.Sprintf(mitLicenseTemplate, currentYear)

	if err := os.WriteFile(licensePath, []byte(licenseText), 0644); err != nil {
		return fmt.Errorf("write LICENSE.txt: %w", err)
	}

	logger.Debug("Generated LICENSE.txt", "path", licensePath)
	return nil
}
