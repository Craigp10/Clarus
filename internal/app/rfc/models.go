package rfc

import "time"

// Document represents a parsed RFC document
type Document struct {
	Path         string
	Title        string
	Summary      string
	Metadata     Metadata
	Sections     []Section
	RawContent   string
	LastModified time.Time
}

// Metadata contains document metadata
type Metadata struct {
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

// SearchResult represents an RFC search result
type SearchResult struct {
	Document   *Document
	Score      float64
	Highlights []Highlight
}

// Highlight represents a search match highlight
type Highlight struct {
	Section string
	Text    string
	Line    int
}

// SearchOptions configures search behavior
type SearchOptions struct {
	MaxResults int
	Category   string
	SortBy     string // relevance, date, title
}

// Index represents the RFC document index
type Index struct {
	Documents   []*Document
	Categories  []string
	LastUpdated time.Time
}

// GlossaryEntry represents a business term definition
type GlossaryEntry struct {
	Term       string
	Definition string
	SourceRFC  string
	Context    string
}
