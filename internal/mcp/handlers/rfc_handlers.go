package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"clarus/internal/app/rfc"
)

// HandleSearchRFCs handles the search_rfcs tool
func (h *Context) HandleSearchRFCs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, ok := request.Params.Arguments["query"].(string)
	if !ok || query == "" {
		return mcp.NewToolResultError("query parameter is required"), nil
	}

	maxResults := 10
	if mr, ok := request.Params.Arguments["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	results, err := h.RFCService.Search(ctx, query, rfc.SearchOptions{
		MaxResults: maxResults,
	})
	if err != nil {
		h.Logger.Error("RFC search failed", "error", err, "query", query)
		return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText("No RFCs found matching your query."), nil
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("# Search Results (%d found)\n\n", len(results)))

	for i, result := range results {
		output.WriteString(fmt.Sprintf("## %d. %s\n", i+1, result.Document.Title))
		output.WriteString(fmt.Sprintf("**Path:** `%s`\n", result.Document.Path))
		output.WriteString(fmt.Sprintf("**Score:** %.2f\n", result.Score))
		if result.Document.Summary != "" {
			output.WriteString(fmt.Sprintf("**Summary:** %s\n", result.Document.Summary))
		}
		if len(result.Highlights) > 0 {
			output.WriteString("**Matches:**\n")
			for _, hl := range result.Highlights {
				output.WriteString(fmt.Sprintf("- Line %d: ...%s...\n", hl.Line, hl.Text))
			}
		}
		output.WriteString("\n")
	}

	return mcp.NewToolResultText(output.String()), nil
}

// HandleReadRFC handles the read_rfc tool
func (h *Context) HandleReadRFC(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok || path == "" {
		return mcp.NewToolResultError("path parameter is required"), nil
	}

	doc, err := h.RFCService.GetDocument(ctx, path)
	if err != nil {
		h.Logger.Error("Failed to read RFC", "error", err, "path", path)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read RFC: %v", err)), nil
	}

	return mcp.NewToolResultText(doc.RawContent), nil
}

// HandleListRFCs handles the list_rfcs tool
func (h *Context) HandleListRFCs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	category := ""
	if cat, ok := request.Params.Arguments["category"].(string); ok {
		category = cat
	}

	docs, err := h.RFCService.ListDocuments(ctx, category)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list RFCs: %v", err)), nil
	}

	if len(docs) == 0 {
		return mcp.NewToolResultText("No RFC documents found."), nil
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("# RFC Documents (%d total)\n\n", len(docs)))

	for _, doc := range docs {
		output.WriteString(fmt.Sprintf("## %s\n", doc.Title))
		output.WriteString(fmt.Sprintf("- **Path:** `%s`\n", doc.Path))
		if doc.Summary != "" {
			output.WriteString(fmt.Sprintf("- **Summary:** %s\n", doc.Summary))
		}
		if doc.Metadata.Status != "" {
			output.WriteString(fmt.Sprintf("- **Status:** %s\n", doc.Metadata.Status))
		}
		output.WriteString(fmt.Sprintf("- **Modified:** %s\n", doc.LastModified.Format("2006-01-02")))
		output.WriteString("\n")
	}

	return mcp.NewToolResultText(output.String()), nil
}

// HandleGetRFCStructure handles the get_rfc_structure tool
func (h *Context) HandleGetRFCStructure(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok || path == "" {
		return mcp.NewToolResultError("path parameter is required"), nil
	}

	sections, err := h.RFCService.GetDocumentStructure(ctx, path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get RFC structure: %v", err)), nil
	}

	var output strings.Builder
	output.WriteString("# Document Structure\n\n")
	formatSections(&output, sections, 0)

	return mcp.NewToolResultText(output.String()), nil
}

func formatSections(output *strings.Builder, sections []rfc.Section, indent int) {
	prefix := strings.Repeat("  ", indent)
	for _, section := range sections {
		output.WriteString(fmt.Sprintf("%s- %s (lines %d-%d)\n", prefix, section.Title, section.StartLine, section.EndLine))
		if len(section.Children) > 0 {
			formatSections(output, section.Children, indent+1)
		}
	}
}

// HandleSearchRFCSections handles the search_rfc_sections tool
func (h *Context) HandleSearchRFCSections(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, ok := request.Params.Arguments["query"].(string)
	if !ok || query == "" {
		return mcp.NewToolResultError("query parameter is required"), nil
	}

	sectionPattern := ""
	if sp, ok := request.Params.Arguments["section_pattern"].(string); ok {
		sectionPattern = sp
	}

	results, err := h.RFCService.SearchSections(ctx, query, sectionPattern)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText("No matching sections found."), nil
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("# Section Search Results (%d found)\n\n", len(results)))

	for i, result := range results {
		output.WriteString(fmt.Sprintf("## %d. %s\n", i+1, result.Document.Title))
		output.WriteString(fmt.Sprintf("**Path:** `%s`\n", result.Document.Path))
		for _, hl := range result.Highlights {
			output.WriteString(fmt.Sprintf("**Section:** %s\n", hl.Section))
			output.WriteString(fmt.Sprintf("```\n%s\n```\n", hl.Text))
		}
		output.WriteString("\n")
	}

	return mcp.NewToolResultText(output.String()), nil
}

// HandleFindRelatedRFCs handles the find_related_rfcs tool
func (h *Context) HandleFindRelatedRFCs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Can search by topic or by RFC path
	topic, hasPath := request.Params.Arguments["rfc_path"].(string)
	if !hasPath {
		topic, _ = request.Params.Arguments["topic"].(string)
	}

	if topic == "" {
		return mcp.NewToolResultError("either topic or rfc_path parameter is required"), nil
	}

	if hasPath {
		related, err := h.RFCService.FindRelated(ctx, topic)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to find related RFCs: %v", err)), nil
		}

		if len(related) == 0 {
			return mcp.NewToolResultText("No related RFCs found."), nil
		}

		var output strings.Builder
		output.WriteString("# Related RFCs\n\n")
		for _, doc := range related {
			output.WriteString(fmt.Sprintf("- **%s** (`%s`)\n", doc.Title, doc.Path))
			if doc.Summary != "" {
				output.WriteString(fmt.Sprintf("  %s\n", doc.Summary))
			}
		}

		return mcp.NewToolResultText(output.String()), nil
	}

	// Search by topic
	results, err := h.RFCService.Search(ctx, topic, rfc.SearchOptions{MaxResults: 10})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText("No RFCs found for this topic."), nil
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("# RFCs Related to '%s'\n\n", topic))
	for _, result := range results {
		output.WriteString(fmt.Sprintf("- **%s** (`%s`) - Score: %.1f\n", result.Document.Title, result.Document.Path, result.Score))
	}

	return mcp.NewToolResultText(output.String()), nil
}

// HandleExplainWorkflow handles the explain_workflow tool
func (h *Context) HandleExplainWorkflow(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workflowName, ok := request.Params.Arguments["workflow_name"].(string)
	if !ok || workflowName == "" {
		return mcp.NewToolResultError("workflow_name parameter is required"), nil
	}

	// Search for workflow-related content in RFCs
	results, err := h.RFCService.Search(ctx, workflowName, rfc.SearchOptions{MaxResults: 5})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No workflow documentation found for '%s'.", workflowName)), nil
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("# Workflow: %s\n\n", workflowName))
	output.WriteString("## Related Documentation\n\n")

	for _, result := range results {
		output.WriteString(fmt.Sprintf("### %s\n", result.Document.Title))
		output.WriteString(fmt.Sprintf("**Source:** `%s`\n\n", result.Document.Path))

		// Include relevant content
		if result.Document.Summary != "" {
			output.WriteString(result.Document.Summary + "\n\n")
		}

		for _, hl := range result.Highlights {
			output.WriteString(fmt.Sprintf("> %s\n\n", hl.Text))
		}
	}

	return mcp.NewToolResultText(output.String()), nil
}

// HandleAnalyzeImpact handles the analyze_impact tool
func (h *Context) HandleAnalyzeImpact(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	description, ok := request.Params.Arguments["description"].(string)
	if !ok || description == "" {
		return mcp.NewToolResultError("description parameter is required"), nil
	}

	// Extract keywords from description
	words := strings.Fields(strings.ToLower(description))
	var keywords []string
	for _, word := range words {
		if len(word) > 3 {
			keywords = append(keywords, word)
		}
	}

	var output strings.Builder
	output.WriteString("# Impact Analysis\n\n")
	output.WriteString(fmt.Sprintf("**Proposed Change:** %s\n\n", description))

	// Search RFCs
	output.WriteString("## Potentially Affected RFCs\n\n")
	for _, kw := range keywords[:min(3, len(keywords))] {
		results, _ := h.RFCService.Search(ctx, kw, rfc.SearchOptions{MaxResults: 3})
		for _, r := range results {
			output.WriteString(fmt.Sprintf("- **%s** (`%s`)\n", r.Document.Title, r.Document.Path))
		}
	}

	// Search repos
	output.WriteString("\n## Potentially Affected Code\n\n")
	repos, _ := h.RepoService.ListRepositories(ctx)
	for _, r := range repos {
		output.WriteString(fmt.Sprintf("- **%s** - %s\n", r.Name, r.Description))
	}

	output.WriteString("\n*Note: This is a preliminary analysis. Review the specific RFCs and code for detailed impact.*\n")

	return mcp.NewToolResultText(output.String()), nil
}

// HandleGetRFCIndex handles the rfc://index resource
func (h *Context) HandleGetRFCIndex(ctx context.Context) (string, error) {
	index, err := h.RFCService.GetIndex(ctx)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(map[string]any{
		"document_count": len(index.Documents),
		"categories":     index.Categories,
		"last_updated":   index.LastUpdated,
		"documents": func() []map[string]any {
			docs := make([]map[string]any, len(index.Documents))
			for i, d := range index.Documents {
				docs[i] = map[string]any{
					"path":    d.Path,
					"title":   d.Title,
					"summary": d.Summary,
					"status":  d.Metadata.Status,
				}
			}
			return docs
		}(),
	}, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
