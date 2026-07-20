package utils

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ReadFileContent reads a file and returns its content as a string
func ReadFileContent(path string, maxSize int64) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Check file size
	info, err := file.Stat()
	if err != nil {
		return "", err
	}

	if info.Size() > maxSize {
		// Read only up to maxSize
		reader := io.LimitReader(file, maxSize)
		content, err := io.ReadAll(reader)
		if err != nil {
			return "", err
		}
		return string(content) + "\n... [truncated]", nil
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// ReadFileLines reads a file and returns lines with optional context
func ReadFileLines(path string, startLine, endLine int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		if lineNum >= startLine && (endLine <= 0 || lineNum <= endLine) {
			lines = append(lines, scanner.Text())
		}
		if endLine > 0 && lineNum > endLine {
			break
		}
	}

	return lines, scanner.Err()
}

// WalkDirectory walks a directory and returns file info for each entry
func WalkDirectory(root string, maxDepth int, excludePatterns []string, includeFiles bool) ([]FileEntry, error) {
	var entries []FileEntry

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip entries with errors
		}

		// Calculate depth
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}

		if relPath == "." {
			return nil
		}

		depth := strings.Count(relPath, string(os.PathSeparator)) + 1

		// Check if we should exclude this entry
		name := d.Name()
		for _, pattern := range excludePatterns {
			if matched, _ := filepath.Match(pattern, name); matched {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Check depth limit
		if maxDepth > 0 && depth > maxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip files if not requested
		if !includeFiles && !d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		entries = append(entries, FileEntry{
			Path:         relPath,
			Name:         name,
			IsDir:        d.IsDir(),
			Size:         info.Size(),
			LastModified: info.ModTime(),
		})

		return nil
	})

	return entries, err
}

// FileEntry represents a file or directory entry
type FileEntry struct {
	Path         string
	Name         string
	IsDir        bool
	Size         int64
	LastModified time.Time
}

// FindFiles finds files matching a glob pattern
func FindFiles(root string, pattern string, excludePatterns []string) ([]string, error) {
	var matches []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// Check exclusions
		name := d.Name()
		for _, excl := range excludePatterns {
			if matched, _ := filepath.Match(excl, name); matched {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if d.IsDir() {
			return nil
		}

		// Check if file matches pattern
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}

		matched, err := filepath.Match(pattern, name)
		if err != nil {
			return nil
		}

		if matched {
			matches = append(matches, relPath)
		}

		return nil
	})

	return matches, err
}

// DetectLanguage attempts to detect the programming language of a file
func DetectLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	languages := map[string]string{
		".go":    "go",
		".js":    "javascript",
		".ts":    "typescript",
		".tsx":   "typescript",
		".jsx":   "javascript",
		".py":    "python",
		".rb":    "ruby",
		".java":  "java",
		".kt":    "kotlin",
		".rs":    "rust",
		".c":     "c",
		".cpp":   "cpp",
		".h":     "c",
		".hpp":   "cpp",
		".cs":    "csharp",
		".swift": "swift",
		".php":   "php",
		".sh":    "shell",
		".bash":  "shell",
		".zsh":   "shell",
		".sql":   "sql",
		".md":    "markdown",
		".yaml":  "yaml",
		".yml":   "yaml",
		".json":  "json",
		".xml":   "xml",
		".html":  "html",
		".css":   "css",
		".scss":  "scss",
		".sass":  "sass",
		".less":  "less",
	}

	if lang, ok := languages[ext]; ok {
		return lang
	}

	return "unknown"
}
