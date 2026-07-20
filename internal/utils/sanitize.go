package utils

import (
	"fmt"
	"path/filepath"
	"strings"
)

// SanitizePath validates and sanitizes a file path to prevent directory traversal
func SanitizePath(basePath, userPath string) (string, error) {
	// Clean both paths
	basePath = filepath.Clean(basePath)
	userPath = filepath.Clean(userPath)

	// If userPath is absolute, make it relative
	if filepath.IsAbs(userPath) {
		// Check if it's under basePath
		if !strings.HasPrefix(userPath, basePath) {
			return "", fmt.Errorf("path is outside allowed directory")
		}
		return userPath, nil
	}

	// Join the paths
	fullPath := filepath.Join(basePath, userPath)
	fullPath = filepath.Clean(fullPath)

	// Verify the result is still under basePath
	if !strings.HasPrefix(fullPath, basePath) {
		return "", fmt.Errorf("path traversal detected: %s", userPath)
	}

	return fullPath, nil
}

// ValidateRepoName validates a repository name
func ValidateRepoName(name string) error {
	if name == "" {
		return fmt.Errorf("repository name cannot be empty")
	}

	// Check for path traversal attempts
	if strings.Contains(name, "..") {
		return fmt.Errorf("repository name cannot contain '..'")
	}

	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("repository name cannot contain path separators")
	}

	// Check for hidden directories
	if strings.HasPrefix(name, ".") {
		return fmt.Errorf("repository name cannot start with '.'")
	}

	return nil
}

// IsSensitiveFile checks if a file might contain sensitive data
func IsSensitiveFile(filename string) bool {
	sensitivePatterns := []string{
		".env",
		".env.*",
		"*.pem",
		"*.key",
		"*.crt",
		"*.p12",
		"*.pfx",
		"id_rsa",
		"id_rsa.*",
		"id_ed25519",
		"id_ed25519.*",
		"credentials",
		"credentials.*",
		"secrets",
		"secrets.*",
		"*.secret",
		"*.secrets",
		"config.local.*",
		"*.local.json",
		"*.local.yaml",
		"*.local.yml",
	}

	base := filepath.Base(filename)
	lower := strings.ToLower(base)

	for _, pattern := range sensitivePatterns {
		if matched, _ := filepath.Match(strings.ToLower(pattern), lower); matched {
			return true
		}
	}

	return false
}

// IsExcluded checks if a path matches any exclusion pattern
func IsExcluded(path string, patterns []string) bool {
	name := filepath.Base(path)

	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
	}

	return false
}
