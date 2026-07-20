package rfc

import (
	"regexp"
	"strings"
	"time"
)

// Parser handles parsing of RFC markdown documents
type Parser struct {
	headingRegex *regexp.Regexp
}

// NewParser creates a new Parser
func NewParser() *Parser {
	return &Parser{
		headingRegex: regexp.MustCompile(`^(#{1,6})\s+(.+)$`),
	}
}

// Parse parses a markdown document into a Document
func (p *Parser) Parse(path, content string, modTime time.Time) (*Document, error) {
	lines := strings.Split(content, "\n")

	doc := &Document{
		Path:         path,
		RawContent:   content,
		LastModified: modTime,
		Sections:     make([]Section, 0),
	}

	// Extract title from first H1 or filename
	doc.Title = p.extractTitle(lines, path)

	// Extract summary from first paragraph
	doc.Summary = p.extractSummary(lines)

	// Extract metadata if present (YAML frontmatter style)
	doc.Metadata = p.extractMetadata(lines)

	// Parse sections
	doc.Sections = p.parseSections(lines)

	return doc, nil
}

func (p *Parser) extractTitle(lines []string, path string) string {
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}

	// Fall back to filename without extension
	parts := strings.Split(path, "/")
	filename := parts[len(parts)-1]
	if idx := strings.LastIndex(filename, "."); idx > 0 {
		return filename[:idx]
	}
	return filename
}

func (p *Parser) extractSummary(lines []string) string {
	inFrontmatter := false
	foundTitle := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip YAML frontmatter
		if trimmed == "---" {
			inFrontmatter = !inFrontmatter
			continue
		}
		if inFrontmatter {
			continue
		}

		// Skip title
		if strings.HasPrefix(trimmed, "# ") {
			foundTitle = true
			continue
		}

		// Skip empty lines after title
		if foundTitle && trimmed == "" {
			continue
		}

		// Skip other headings
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Return first non-empty paragraph
		if trimmed != "" && foundTitle {
			// Limit summary length
			if len(trimmed) > 300 {
				return trimmed[:300] + "..."
			}
			return trimmed
		}
	}

	return ""
}

func (p *Parser) extractMetadata(lines []string) Metadata {
	meta := Metadata{}

	// Look for YAML frontmatter
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return meta
	}

	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "---" {
			break
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch key {
		case "author":
			meta.Author = value
		case "status":
			meta.Status = value
		case "date":
			if t, err := time.Parse("2006-01-02", value); err == nil {
				meta.Date = t
			}
		case "tags":
			// Parse comma-separated tags
			tags := strings.Split(value, ",")
			for _, tag := range tags {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					meta.Tags = append(meta.Tags, tag)
				}
			}
		case "references":
			refs := strings.Split(value, ",")
			for _, ref := range refs {
				ref = strings.TrimSpace(ref)
				if ref != "" {
					meta.References = append(meta.References, ref)
				}
			}
		}
	}

	return meta
}

func (p *Parser) parseSections(lines []string) []Section {
	var sections []Section
	var currentSection *Section
	var contentBuilder strings.Builder

	inFrontmatter := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle frontmatter
		if trimmed == "---" {
			inFrontmatter = !inFrontmatter
			continue
		}
		if inFrontmatter {
			continue
		}

		// Check for heading
		matches := p.headingRegex.FindStringSubmatch(line)
		if matches != nil {
			// Save previous section's content
			if currentSection != nil {
				currentSection.Content = strings.TrimSpace(contentBuilder.String())
				currentSection.EndLine = i
			}

			level := len(matches[1])
			title := matches[2]

			newSection := Section{
				Level:     level,
				Title:     title,
				StartLine: i + 1,
				Children:  make([]Section, 0),
			}

			// Add to appropriate place in hierarchy
			if level == 1 || len(sections) == 0 {
				sections = append(sections, newSection)
				currentSection = &sections[len(sections)-1]
			} else {
				// Find parent section at lower level
				parent := findParentSection(sections, level)
				if parent != nil {
					parent.Children = append(parent.Children, newSection)
					currentSection = &parent.Children[len(parent.Children)-1]
				} else {
					sections = append(sections, newSection)
					currentSection = &sections[len(sections)-1]
				}
			}

			contentBuilder.Reset()
			continue
		}

		// Add line to current section content
		if currentSection != nil {
			contentBuilder.WriteString(line)
			contentBuilder.WriteString("\n")
		}
	}

	// Finalize last section
	if currentSection != nil {
		currentSection.Content = strings.TrimSpace(contentBuilder.String())
		currentSection.EndLine = len(lines)
	}

	return sections
}

func findParentSection(sections []Section, level int) *Section {
	for i := len(sections) - 1; i >= 0; i-- {
		if sections[i].Level < level {
			return &sections[i]
		}
		// Check children recursively
		if parent := findParentInChildren(&sections[i], level); parent != nil {
			return parent
		}
	}
	return nil
}

func findParentInChildren(section *Section, level int) *Section {
	for i := len(section.Children) - 1; i >= 0; i-- {
		if section.Children[i].Level < level {
			return &section.Children[i]
		}
		if parent := findParentInChildren(&section.Children[i], level); parent != nil {
			return parent
		}
	}
	if section.Level < level {
		return section
	}
	return nil
}
