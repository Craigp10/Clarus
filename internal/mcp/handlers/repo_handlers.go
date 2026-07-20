package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"clarus/internal/app/repo"
)

// HandleListRepos handles the list_repos tool
func (h *Context) HandleListRepos(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repos, err := h.RepoService.ListRepositories(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list repositories: %v", err)), nil
	}

	if len(repos) == 0 {
		return mcp.NewToolResultText("No repositories found."), nil
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("# Repositories (%d found)\n\n", len(repos)))

	for _, r := range repos {
		output.WriteString(fmt.Sprintf("## %s\n", r.Name))
		if r.Description != "" {
			output.WriteString(fmt.Sprintf("%s\n", r.Description))
		}
		output.WriteString(fmt.Sprintf("- **Language:** %s\n", r.PrimaryLanguage))
		output.WriteString(fmt.Sprintf("- **Files:** %d\n", r.FileCount))
		output.WriteString(fmt.Sprintf("- **Last Modified:** %s\n", r.LastModified.Format("2006-01-02")))
		output.WriteString("\n")
	}

	return mcp.NewToolResultText(output.String()), nil
}

// HandleExploreRepoStructure handles the explore_repo_structure tool
func (h *Context) HandleExploreRepoStructure(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoName, ok := request.Params.Arguments["repo"].(string)
	if !ok || repoName == "" {
		return mcp.NewToolResultError("repo parameter is required"), nil
	}

	maxDepth := 3
	if md, ok := request.Params.Arguments["max_depth"].(float64); ok {
		maxDepth = int(md)
	}

	includeFiles := true
	if incl, ok := request.Params.Arguments["include_files"].(bool); ok {
		includeFiles = incl
	}

	tree, err := h.RepoService.GetDirectoryTree(ctx, repoName, repo.TreeOptions{
		MaxDepth:     maxDepth,
		IncludeFiles: includeFiles,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get directory tree: %v", err)), nil
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("# Repository: %s\n\n", repoName))
	output.WriteString("```\n")
	formatTree(&output, tree.Root, "", true)
	output.WriteString("```\n")

	return mcp.NewToolResultText(output.String()), nil
}

func formatTree(output *strings.Builder, node *TreeNode, prefix string, isLast bool) {
	if node == nil {
		return
	}

	// Print current node
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	if node.Path != "" { // Skip root
		icon := "📁"
		if !node.IsDir {
			icon = "📄"
		}
		output.WriteString(fmt.Sprintf("%s%s%s %s\n", prefix, connector, icon, node.Name))
	}

	// Prepare prefix for children
	childPrefix := prefix
	if node.Path != "" {
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}
	}

	// Print children
	for i, child := range node.Children {
		formatTree(output, child, childPrefix, i == len(node.Children)-1)
	}
}

// TreeNode is imported from repo package but we need local reference
type TreeNode = repo.TreeNode

// HandleReadFile handles the read_file tool
func (h *Context) HandleReadFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoName, ok := request.Params.Arguments["repo"].(string)
	if !ok || repoName == "" {
		return mcp.NewToolResultError("repo parameter is required"), nil
	}

	path, ok := request.Params.Arguments["path"].(string)
	if !ok || path == "" {
		return mcp.NewToolResultError("path parameter is required"), nil
	}

	content, err := h.RepoService.ReadFile(ctx, repoName, path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read file: %v", err)), nil
	}

	return mcp.NewToolResultText(content), nil
}

// HandleSearchCode handles the search_code tool
func (h *Context) HandleSearchCode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, ok := request.Params.Arguments["query"].(string)
	if !ok || query == "" {
		return mcp.NewToolResultError("query parameter is required"), nil
	}

	opts := repo.CodeSearchOptions{
		MaxResults: 50,
		Context:    2,
	}

	if repoName, ok := request.Params.Arguments["repo"].(string); ok {
		opts.Repository = repoName
	}

	if pattern, ok := request.Params.Arguments["file_pattern"].(string); ok {
		opts.FilePattern = pattern
	}

	results, err := h.RepoService.SearchCode(ctx, query, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText("No matches found."), nil
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("# Code Search Results (%d matches)\n\n", len(results)))

	for _, result := range results {
		output.WriteString(fmt.Sprintf("## %s:%s:%d\n", result.Repository, result.FilePath, result.Line))
		output.WriteString("```\n")
		for _, ctx := range result.Context[:len(result.Context)/2] {
			output.WriteString(fmt.Sprintf("  %s\n", ctx))
		}
		output.WriteString(fmt.Sprintf("> %s\n", result.Content))
		for _, ctx := range result.Context[len(result.Context)/2:] {
			output.WriteString(fmt.Sprintf("  %s\n", ctx))
		}
		output.WriteString("```\n\n")
	}

	return mcp.NewToolResultText(output.String()), nil
}

// HandleGetFileOutline handles the get_file_outline tool
func (h *Context) HandleGetFileOutline(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoName, ok := request.Params.Arguments["repo"].(string)
	if !ok || repoName == "" {
		return mcp.NewToolResultError("repo parameter is required"), nil
	}

	path, ok := request.Params.Arguments["path"].(string)
	if !ok || path == "" {
		return mcp.NewToolResultError("path parameter is required"), nil
	}

	content, err := h.RepoService.ReadFile(ctx, repoName, path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read file: %v", err)), nil
	}

	// Simple outline extraction based on language patterns
	outline := extractOutline(content, path)

	var output strings.Builder
	output.WriteString(fmt.Sprintf("# File Outline: %s/%s\n\n", repoName, path))
	output.WriteString(outline)

	return mcp.NewToolResultText(output.String()), nil
}

func extractOutline(content, path string) string {
	lines := strings.Split(content, "\n")
	var outline strings.Builder

	// Detect language
	lang := detectLanguageFromPath(path)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		switch lang {
		case "go":
			if strings.HasPrefix(trimmed, "func ") || strings.HasPrefix(trimmed, "type ") {
				outline.WriteString(fmt.Sprintf("Line %d: %s\n", i+1, trimmed))
			}
		case "python":
			if strings.HasPrefix(trimmed, "def ") || strings.HasPrefix(trimmed, "class ") {
				outline.WriteString(fmt.Sprintf("Line %d: %s\n", i+1, trimmed))
			}
		case "javascript", "typescript":
			if strings.HasPrefix(trimmed, "function ") || strings.HasPrefix(trimmed, "class ") ||
				strings.HasPrefix(trimmed, "const ") || strings.HasPrefix(trimmed, "export ") {
				outline.WriteString(fmt.Sprintf("Line %d: %s\n", i+1, trimmed))
			}
		default:
			// Generic: look for function-like patterns
			if strings.Contains(trimmed, "function") || strings.Contains(trimmed, "class") {
				outline.WriteString(fmt.Sprintf("Line %d: %s\n", i+1, trimmed))
			}
		}
	}

	if outline.Len() == 0 {
		return "No outline available for this file type."
	}

	return outline.String()
}

func detectLanguageFromPath(path string) string {
	if strings.HasSuffix(path, ".go") {
		return "go"
	}
	if strings.HasSuffix(path, ".py") {
		return "python"
	}
	if strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".jsx") {
		return "javascript"
	}
	if strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".tsx") {
		return "typescript"
	}
	return "unknown"
}

// HandleFindFiles handles the find_files tool
func (h *Context) HandleFindFiles(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pattern, ok := request.Params.Arguments["pattern"].(string)
	if !ok || pattern == "" {
		return mcp.NewToolResultError("pattern parameter is required"), nil
	}

	repoName := ""
	if rn, ok := request.Params.Arguments["repo"].(string); ok {
		repoName = rn
	}

	files, err := h.RepoService.FindFiles(ctx, pattern, repoName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Find failed: %v", err)), nil
	}

	if len(files) == 0 {
		return mcp.NewToolResultText("No files found matching pattern."), nil
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("# Files Matching '%s' (%d found)\n\n", pattern, len(files)))

	for _, f := range files {
		output.WriteString(fmt.Sprintf("- `%s` (%d bytes, %s)\n", f.Path, f.Size, f.Language))
	}

	return mcp.NewToolResultText(output.String()), nil
}

// HandleGetRepoIndex handles the repo://index resource
func (h *Context) HandleGetRepoIndex(ctx context.Context) (string, error) {
	repos, err := h.RepoService.ListRepositories(ctx)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(map[string]any{
		"repository_count": len(repos),
		"repositories": func() []map[string]any {
			result := make([]map[string]any, len(repos))
			for i, r := range repos {
				result[i] = map[string]any{
					"name":        r.Name,
					"description": r.Description,
					"language":    r.PrimaryLanguage,
					"file_count":  r.FileCount,
				}
			}
			return result
		}(),
	}, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}
