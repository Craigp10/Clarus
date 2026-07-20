package config

import (
	"fmt"
	"os"
)

// Validate checks that the configuration is valid
func (c *Config) Validate() error {
	// Check RFC_DOCS_ROOT exists and is a directory
	info, err := os.Stat(c.RFCDocs.Root)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("RFC_DOCS_ROOT directory does not exist: %s", c.RFCDocs.Root)
		}
		return fmt.Errorf("RFC_DOCS_ROOT path error: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("RFC_DOCS_ROOT is not a directory: %s", c.RFCDocs.Root)
	}

	// Check REPOS_ROOT exists and is a directory
	info, err = os.Stat(c.Repos.Root)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("REPOS_ROOT directory does not exist: %s", c.Repos.Root)
		}
		return fmt.Errorf("REPOS_ROOT path error: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("REPOS_ROOT is not a directory: %s", c.Repos.Root)
	}

	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", c.Logging.Level)
	}

	// Validate log format
	validFormats := map[string]bool{
		"text": true,
		"json": true,
	}
	if !validFormats[c.Logging.Format] {
		return fmt.Errorf("invalid log format: %s (must be text or json)", c.Logging.Format)
	}

	// Validate max file size is positive
	if c.Repos.MaxFileSize <= 0 {
		return fmt.Errorf("CLARUS_MAX_FILE_SIZE must be positive")
	}

	// Validate file extensions start with a dot
	for _, ext := range c.RFCDocs.FileExtensions {
		if len(ext) == 0 || ext[0] != '.' {
			return fmt.Errorf("file extension must start with a dot: %s", ext)
		}
	}

	return nil
}
