package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Name:    "business-context-agent",
			Version: "1.0.0",
		},
		RFCDocs: RFCDocsConfig{
			FileExtensions: []string{".md", ".markdown"},
		},
		Repos: ReposConfig{
			MaxFileSize:     10 * 1024 * 1024, // 10MB
			ExcludePatterns: []string{".git", "node_modules", "vendor", "__pycache__", ".venv", "target", "build", "dist"},
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}

	// Required variables
	rfcRoot := os.Getenv("RFC_DOCS_ROOT")
	if rfcRoot == "" {
		return nil, fmt.Errorf("RFC_DOCS_ROOT environment variable is required")
	}
	cfg.RFCDocs.Root = rfcRoot

	reposRoot := os.Getenv("REPOS_ROOT")
	if reposRoot == "" {
		return nil, fmt.Errorf("REPOS_ROOT environment variable is required")
	}
	cfg.Repos.Root = reposRoot

	// Optional variables
	if level := os.Getenv("CLARUS_LOG_LEVEL"); level != "" {
		cfg.Logging.Level = level
	}

	if format := os.Getenv("CLARUS_LOG_FORMAT"); format != "" {
		cfg.Logging.Format = format
	}

	if maxSize := os.Getenv("CLARUS_MAX_FILE_SIZE"); maxSize != "" {
		size, err := strconv.ParseInt(maxSize, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid CLARUS_MAX_FILE_SIZE: %w", err)
		}
		cfg.Repos.MaxFileSize = size
	}

	if patterns := os.Getenv("CLARUS_EXCLUDE_PATTERNS"); patterns != "" {
		cfg.Repos.ExcludePatterns = strings.Split(patterns, ",")
		for i, p := range cfg.Repos.ExcludePatterns {
			cfg.Repos.ExcludePatterns[i] = strings.TrimSpace(p)
		}
	}

	if extensions := os.Getenv("CLARUS_RFC_EXTENSIONS"); extensions != "" {
		cfg.RFCDocs.FileExtensions = strings.Split(extensions, ",")
		for i, e := range cfg.RFCDocs.FileExtensions {
			cfg.RFCDocs.FileExtensions[i] = strings.TrimSpace(e)
		}
	}

	return cfg, nil
}
