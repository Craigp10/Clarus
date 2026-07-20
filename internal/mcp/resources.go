package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"

	"clarus/internal/app/repo"
)

// registerResources registers all MCP resources with the server
func (s *Server) registerResources() {
	// RFC Index - static resource
	s.mcpServer.AddResource(
		mcp.NewResource(
			"rfc://index",
			"RFC Index",
			mcp.WithResourceDescription("Complete index of all RFC documents with metadata"),
			mcp.WithMIMEType("application/json"),
		),
		func(ctx context.Context, request mcp.ReadResourceRequest) (string, error) {
			return s.handlers.HandleGetRFCIndex(ctx)
		},
	)

	// Repository Index - static resource
	s.mcpServer.AddResource(
		mcp.NewResource(
			"repo://index",
			"Repository Index",
			mcp.WithResourceDescription("List of all repositories with metadata"),
			mcp.WithMIMEType("application/json"),
		),
		func(ctx context.Context, request mcp.ReadResourceRequest) (string, error) {
			return s.handlers.HandleGetRepoIndex(ctx)
		},
	)

	// RFC Document Template - dynamic resource
	s.mcpServer.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"rfc://{path}",
			"RFC Document",
			mcp.WithTemplateDescription("Content of a specific RFC document"),
			mcp.WithTemplateMIMEType("text/markdown"),
		),
		func(ctx context.Context, request mcp.ReadResourceRequest) (string, error) {
			// Extract path from URI: rfc://path/to/doc.md -> path/to/doc.md
			path := request.Params.URI[6:] // Remove "rfc://"
			doc, err := s.handlers.RFCService.GetDocument(ctx, path)
			if err != nil {
				return "", err
			}
			return doc.RawContent, nil
		},
	)

	// Repository Tree Template - dynamic resource
	s.mcpServer.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"repo://{name}/tree",
			"Repository Tree",
			mcp.WithTemplateDescription("Directory structure of a repository"),
			mcp.WithTemplateMIMEType("application/json"),
		),
		func(ctx context.Context, request mcp.ReadResourceRequest) (string, error) {
			// Extract repo name from URI: repo://name/tree -> name
			uri := request.Params.URI
			// Remove "repo://" prefix and "/tree" suffix
			name := uri[7 : len(uri)-5]

			tree, err := s.handlers.RepoService.GetDirectoryTree(ctx, name, repo.TreeOptions{
				MaxDepth:     5,
				IncludeFiles: true,
			})
			if err != nil {
				return "", err
			}

			return formatTreeJSON(tree.Root), nil
		},
	)

	// Repository File Template - dynamic resource
	s.mcpServer.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"repo://{name}/file/{path}",
			"Repository File",
			mcp.WithTemplateDescription("Content of a file in a repository"),
			mcp.WithTemplateMIMEType("text/plain"),
		),
		func(ctx context.Context, request mcp.ReadResourceRequest) (string, error) {
			// Parse URI: repo://name/file/path/to/file.go
			uri := request.Params.URI[7:] // Remove "repo://"
			// Find /file/ separator
			idx := 0
			for i := 0; i < len(uri); i++ {
				if uri[i:i+6] == "/file/" {
					idx = i
					break
				}
			}
			if idx == 0 {
				return "", nil
			}

			repoName := uri[:idx]
			filePath := uri[idx+6:]

			return s.handlers.RepoService.ReadFile(ctx, repoName, filePath)
		},
	)
}

func formatTreeJSON(node *repo.TreeNode) string {
	if node == nil {
		return "{}"
	}

	return formatNodeJSON(node)
}

func formatNodeJSON(node *repo.TreeNode) string {
	if node == nil {
		return "null"
	}

	result := `{"name":"` + escapeJSON(node.Name) + `","path":"` + escapeJSON(node.Path) + `","isDir":` + boolStr(node.IsDir)

	if len(node.Children) > 0 {
		result += `,"children":[`
		for i, child := range node.Children {
			if i > 0 {
				result += ","
			}
			result += formatNodeJSON(child)
		}
		result += `]`
	}

	result += `}`
	return result
}

func escapeJSON(s string) string {
	// Simple JSON string escaping
	result := ""
	for _, c := range s {
		switch c {
		case '"':
			result += `\"`
		case '\\':
			result += `\\`
		case '\n':
			result += `\n`
		case '\r':
			result += `\r`
		case '\t':
			result += `\t`
		default:
			result += string(c)
		}
	}
	return result
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
