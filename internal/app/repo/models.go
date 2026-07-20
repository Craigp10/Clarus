package repo

import "time"

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

// CodeSearchResult represents a code search match
type CodeSearchResult struct {
	Repository string
	FilePath   string
	Line       int
	Content    string
	Context    []string
}

// TreeOptions configures directory tree generation
type TreeOptions struct {
	MaxDepth        int
	IncludeFiles    bool
	ExcludePatterns []string
}

// CodeSearchOptions configures code search
type CodeSearchOptions struct {
	Repository  string
	FilePattern string
	MaxResults  int
	Context     int
}
