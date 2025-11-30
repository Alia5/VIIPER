package rust

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

const errorTemplate = `
use std::fmt;

/// Errors that can occur when using the VIIPER client
#[derive(Debug, thiserror::Error)]
pub enum ViiperError {
    /// Network or I/O errors
    #[error("transport error: {0}")]
    Io(#[from] std::io::Error),
    
    /// RFC 7807 Problem+JSON response from server
    #[error("{0}")]
    Protocol(#[from] ProblemJson),
    
    /// Failed to parse JSON response
    #[error("parse error: {0}")]
    Parse(#[from] serde_json::Error),
    
    /// Unexpected response format
    #[error("unexpected response: {0}")]
    UnexpectedResponse(String),
    
    /// Operation timed out
    #[cfg(feature = "async")]
    #[error("operation timed out")]
    Timeout,
}

/// RFC 7807 Problem Details for HTTP APIs
#[derive(Debug, Clone, serde::Deserialize)]
pub struct ProblemJson {
    pub status: u16,
    pub title: String,
    pub detail: String,
}

impl fmt::Display for ProblemJson {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{} {}: {}", self.status, self.title, self.detail)
    }
}

impl std::error::Error for ProblemJson {}
`

func generateError(logger *slog.Logger, srcDir string) error {
	logger.Debug("Generating error.rs")
	outputFile := filepath.Join(srcDir, "error.rs")

	content := writeFileHeaderRust() + errorTemplate

	if err := os.WriteFile(outputFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("write error.rs: %w", err)
	}

	logger.Info("Generated error.rs", "file", outputFile)
	return nil
}
