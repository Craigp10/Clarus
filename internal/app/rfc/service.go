package rfc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"clarus/internal/config"
	"clarus/internal/utils"
)

// Service defines the RFC document operations
type Service interface {
	ListDocuments(ctx context.Context, category string) ([]*Document, error)
	GetDocument(ctx context.Context, path string) (*Document, error)
	GetDocumentStructure(ctx context.Context, path string) ([]Section, error)
	GetSection(ctx context.Context, path string, sectionName string) (*Section, error)
	Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error)
	SearchSections(ctx context.Context, query string, sectionPattern string) ([]SearchResult, error)
	RefreshIndex(ctx context.Context) error
	GetIndex(ctx context.Context) (*Index, error)
	FindRelated(ctx context.Context, path string) ([]*Document, error)
}

// serviceImpl implements the Service interface
type serviceImpl struct {
	cfg    *config.RFCDocsConfig
	parser *Parser
	index  *Index
	mu     sync.RWMutex
}

// NewService creates a new RFC service
func NewService(cfg *config.RFCDocsConfig) (Service, error) {
	s := &serviceImpl{
		cfg:    cfg,
		parser: NewParser(),
		index:  &Index{Documents: make([]*Document, 0)},
	}

	// Build initial index
	if err := s.RefreshIndex(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to build initial index: %w", err)
	}

	return s, nil
}

// ListDocuments returns all RFC documents, optionally filtered by category
func (s *serviceImpl) ListDocuments(ctx context.Context, category string) ([]*Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if category == "" {
		return s.index.Documents, nil
	}

	var filtered []*Document
	for _, doc := range s.index.Documents {
		if strings.HasPrefix(doc.Path, category) {
			filtered = append(filtered, doc)
		}
	}

	return filtered, nil
}

// GetDocument returns a specific RFC document by path
func (s *serviceImpl) GetDocument(ctx context.Context, path string) (*Document, error) {
	fullPath, err := utils.SanitizePath(s.cfg.Root, path)
	if err != nil {
		return nil, err
	}

	content, err := utils.ReadFileContent(fullPath, 10*1024*1024) // 10MB limit
	if err != nil {
		return nil, fmt.Errorf("failed to read RFC: %w", err)
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	return s.parser.Parse(path, content, info.ModTime())
}

// GetDocumentStructure returns the section structure of an RFC
func (s *serviceImpl) GetDocumentStructure(ctx context.Context, path string) ([]Section, error) {
	doc, err := s.GetDocument(ctx, path)
	if err != nil {
		return nil, err
	}
	return doc.Sections, nil
}

// GetSection returns a specific section from an RFC
func (s *serviceImpl) GetSection(ctx context.Context, path string, sectionName string) (*Section, error) {
	doc, err := s.GetDocument(ctx, path)
	if err != nil {
		return nil, err
	}

	section := findSection(doc.Sections, sectionName)
	if section == nil {
		return nil, fmt.Errorf("section not found: %s", sectionName)
	}

	return section, nil
}

func findSection(sections []Section, name string) *Section {
	name = strings.ToLower(name)
	for i := range sections {
		if strings.ToLower(sections[i].Title) == name ||
			strings.Contains(strings.ToLower(sections[i].Title), name) {
			return &sections[i]
		}
		if found := findSection(sections[i].Children, name); found != nil {
			return found
		}
	}
	return nil
}

// Search performs full-text search across all RFC documents
func (s *serviceImpl) Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if opts.MaxResults <= 0 {
		opts.MaxResults = 10
	}

	query = strings.ToLower(query)
	var results []SearchResult

	for _, doc := range s.index.Documents {
		// Filter by category if specified
		if opts.Category != "" && !strings.HasPrefix(doc.Path, opts.Category) {
			continue
		}

		score, highlights := s.scoreDocument(doc, query)
		if score > 0 {
			results = append(results, SearchResult{
				Document:   doc,
				Score:      score,
				Highlights: highlights,
			})
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limit results
	if len(results) > opts.MaxResults {
		results = results[:opts.MaxResults]
	}

	return results, nil
}

func (s *serviceImpl) scoreDocument(doc *Document, query string) (float64, []Highlight) {
	var score float64
	var highlights []Highlight

	content := strings.ToLower(doc.RawContent)
	title := strings.ToLower(doc.Title)

	// Title match is worth more
	if strings.Contains(title, query) {
		score += 10
		highlights = append(highlights, Highlight{
			Section: "title",
			Text:    doc.Title,
			Line:    1,
		})
	}

	// Count content matches
	matches := strings.Count(content, query)
	if matches > 0 {
		score += float64(matches)

		// Find context for first few matches
		lines := strings.Split(doc.RawContent, "\n")
		matchCount := 0
		for i, line := range lines {
			if strings.Contains(strings.ToLower(line), query) && matchCount < 3 {
				highlights = append(highlights, Highlight{
					Section: "",
					Text:    truncateString(line, 100),
					Line:    i + 1,
				})
				matchCount++
			}
		}
	}

	return score, highlights
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// SearchSections searches within specific sections of RFCs
func (s *serviceImpl) SearchSections(ctx context.Context, query string, sectionPattern string) ([]SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query = strings.ToLower(query)
	sectionPattern = strings.ToLower(sectionPattern)

	var results []SearchResult

	for _, doc := range s.index.Documents {
		for _, section := range doc.Sections {
			if !strings.Contains(strings.ToLower(section.Title), sectionPattern) {
				continue
			}

			content := strings.ToLower(section.Content)
			if strings.Contains(content, query) {
				results = append(results, SearchResult{
					Document: doc,
					Score:    float64(strings.Count(content, query)),
					Highlights: []Highlight{{
						Section: section.Title,
						Text:    truncateString(section.Content, 200),
						Line:    section.StartLine,
					}},
				})
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// RefreshIndex rebuilds the document index
func (s *serviceImpl) RefreshIndex(ctx context.Context) error {
	var docs []*Document
	categories := make(map[string]bool)

	err := filepath.WalkDir(s.cfg.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if d.IsDir() {
			return nil
		}

		// Check file extension
		ext := filepath.Ext(path)
		isValidExt := false
		for _, validExt := range s.cfg.FileExtensions {
			if ext == validExt {
				isValidExt = true
				break
			}
		}
		if !isValidExt {
			return nil
		}

		// Check exclusions
		if utils.IsExcluded(path, s.cfg.ExcludePatterns) {
			return nil
		}

		relPath, err := filepath.Rel(s.cfg.Root, path)
		if err != nil {
			return nil
		}

		content, err := utils.ReadFileContent(path, 10*1024*1024)
		if err != nil {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		doc, err := s.parser.Parse(relPath, content, info.ModTime())
		if err != nil {
			return nil
		}

		docs = append(docs, doc)

		// Track categories (top-level directories)
		parts := strings.Split(relPath, string(os.PathSeparator))
		if len(parts) > 1 {
			categories[parts[0]] = true
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Sort documents by path
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].Path < docs[j].Path
	})

	// Build category list
	var cats []string
	for cat := range categories {
		cats = append(cats, cat)
	}
	sort.Strings(cats)

	s.mu.Lock()
	s.index = &Index{
		Documents:   docs,
		Categories:  cats,
		LastUpdated: time.Now(),
	}
	s.mu.Unlock()

	return nil
}

// GetIndex returns the current document index
func (s *serviceImpl) GetIndex(ctx context.Context) (*Index, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.index, nil
}

// FindRelated finds RFCs related to the given document
func (s *serviceImpl) FindRelated(ctx context.Context, path string) ([]*Document, error) {
	doc, err := s.GetDocument(ctx, path)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Simple keyword-based matching
	keywords := extractKeywords(doc.Title + " " + doc.Summary)

	var scored []struct {
		doc   *Document
		score float64
	}

	for _, other := range s.index.Documents {
		if other.Path == path {
			continue
		}

		score := 0.0
		otherText := strings.ToLower(other.Title + " " + other.Summary + " " + other.RawContent)

		for _, kw := range keywords {
			if strings.Contains(otherText, kw) {
				score++
			}
		}

		if score > 0 {
			scored = append(scored, struct {
				doc   *Document
				score float64
			}{other, score})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	var related []*Document
	for i := 0; i < len(scored) && i < 5; i++ {
		related = append(related, scored[i].doc)
	}

	return related, nil
}

func extractKeywords(text string) []string {
	// Simple keyword extraction - split and filter common words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "from": true,
		"is": true, "are": true, "was": true, "were": true, "be": true,
		"been": true, "being": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "will": true, "would": true,
		"could": true, "should": true, "may": true, "might": true, "must": true,
		"this": true, "that": true, "these": true, "those": true,
	}

	words := strings.Fields(strings.ToLower(text))
	var keywords []string
	seen := make(map[string]bool)

	for _, word := range words {
		word = strings.Trim(word, ".,!?;:\"'()[]{}")
		if len(word) > 3 && !stopWords[word] && !seen[word] {
			keywords = append(keywords, word)
			seen[word] = true
		}
	}

	return keywords
}
