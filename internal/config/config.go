package config

import "time"

// Config holds all application configuration
type Config struct {
	Server  ServerConfig
	RFCDocs RFCDocsConfig
	Repos   ReposConfig
	Logging LoggingConfig
}

// ServerConfig holds MCP server configuration
type ServerConfig struct {
	Name    string
	Version string
}

// RFCDocsConfig holds RFC documents configuration
type RFCDocsConfig struct {
	Root            string
	FileExtensions  []string
	ExcludePatterns []string
}

// ReposConfig holds repository configuration
type ReposConfig struct {
	Root            string
	ExcludePatterns []string
	MaxFileSize     int64
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// RFCDocument represents a parsed RFC document
type RFCDocument struct {
	Path         string
	Title        string
	Summary      string
	Metadata     RFCMetadata
	Sections     []Section
	RawContent   string
	LastModified time.Time
}

// RFCMetadata contains document metadata
type RFCMetadata struct {
	Author     string
	Date       time.Time
	Status     string
	Tags       []string
	References []string
}

// Section represents a document section
type Section struct {
	Level     int
	Title     string
	Content   string
	Children  []Section
	StartLine int
	EndLine   int
}

// Repository represents a code repository
type Repository struct {
	Name            string
	Path            string
	Description     string
	LastModified    time.Time
	PrimaryLanguage string
	FileCount       int
}

// FileInfo represents a file in a repository
type FileInfo struct {
	Path         string
	Name         string
	IsDirectory  bool
	Size         int64
	LastModified time.Time
	Language     string
}

// DirectoryTree represents a repository's structure
type DirectoryTree struct {
	Root     *TreeNode
	MaxDepth int
}

// TreeNode represents a node in the directory tree
type TreeNode struct {
	Name     string
	Path     string
	IsDir    bool
	Children []*TreeNode
}

// SearchResult represents a search result
type SearchResult struct {
	Path       string
	Title      string
	Score      float64
	Highlights []Highlight
	Summary    string
}

// Highlight represents a search match highlight
type Highlight struct {
	Section string
	Text    string
	Line    int
}

// CodeSearchResult represents a code search match
type CodeSearchResult struct {
	Repository string
	FilePath   string
	Line       int
	Content    string
	Context    []string
}
