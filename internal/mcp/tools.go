package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// registerTools registers all MCP tools with the server
func (s *Server) registerTools() {
	// RFC Tools
	s.mcpServer.AddTool(
		mcp.NewTool("search_rfcs",
			mcp.WithDescription("Full-text search across all RFC documents. Returns matching RFCs with relevance scores and highlighted matches."),
			mcp.WithString("query",
				mcp.Description("Search term or phrase to find in RFC documents"),
				mcp.Required(),
			),
			mcp.WithNumber("max_results",
				mcp.Description("Maximum number of results to return (default: 10)"),
			),
		),
		s.handlers.HandleSearchRFCs,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("read_rfc",
			mcp.WithDescription("Read the complete content of a specific RFC document by its path."),
			mcp.WithString("path",
				mcp.Description("Relative path to the RFC file from RFC_DOCS_ROOT"),
				mcp.Required(),
			),
		),
		s.handlers.HandleReadRFC,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("list_rfcs",
			mcp.WithDescription("List all available RFC documents with their titles, summaries, and metadata."),
			mcp.WithString("category",
				mcp.Description("Filter by category/subdirectory (optional)"),
			),
		),
		s.handlers.HandleListRFCs,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("get_rfc_structure",
			mcp.WithDescription("Extract the heading structure and sections of an RFC document."),
			mcp.WithString("path",
				mcp.Description("Relative path to the RFC file"),
				mcp.Required(),
			),
		),
		s.handlers.HandleGetRFCStructure,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("search_rfc_sections",
			mcp.WithDescription("Search within specific sections of RFCs (e.g., only 'Requirements' or 'Decisions' sections)."),
			mcp.WithString("query",
				mcp.Description("Search term to find"),
				mcp.Required(),
			),
			mcp.WithString("section_pattern",
				mcp.Description("Pattern to match section names (e.g., 'requirement', 'decision')"),
			),
		),
		s.handlers.HandleSearchRFCSections,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("find_related_rfcs",
			mcp.WithDescription("Find RFCs related to a specific topic or another RFC document."),
			mcp.WithString("topic",
				mcp.Description("Topic keyword to search for"),
			),
			mcp.WithString("rfc_path",
				mcp.Description("Path to source RFC to find related documents for"),
			),
		),
		s.handlers.HandleFindRelatedRFCs,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("explain_workflow",
			mcp.WithDescription("Extract and explain a workflow from RFC documents based on a workflow name or keyword."),
			mcp.WithString("workflow_name",
				mcp.Description("Name or keyword for the workflow to explain"),
				mcp.Required(),
			),
		),
		s.handlers.HandleExplainWorkflow,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("analyze_impact",
			mcp.WithDescription("Analyze which RFCs and repositories might be affected by a proposed change."),
			mcp.WithString("description",
				mcp.Description("Description of the proposed change"),
				mcp.Required(),
			),
		),
		s.handlers.HandleAnalyzeImpact,
	)

	// Repository Tools
	s.mcpServer.AddTool(
		mcp.NewTool("list_repos",
			mcp.WithDescription("List all repositories under REPOS_ROOT with their metadata including primary language and file count."),
		),
		s.handlers.HandleListRepos,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("explore_repo_structure",
			mcp.WithDescription("Get the directory tree structure of a repository."),
			mcp.WithString("repo",
				mcp.Description("Repository name"),
				mcp.Required(),
			),
			mcp.WithNumber("max_depth",
				mcp.Description("Maximum directory depth to traverse (default: 3)"),
			),
			mcp.WithBoolean("include_files",
				mcp.Description("Include files in the tree, not just directories (default: true)"),
			),
		),
		s.handlers.HandleExploreRepoStructure,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("read_file",
			mcp.WithDescription("Read a specific file from a repository."),
			mcp.WithString("repo",
				mcp.Description("Repository name"),
				mcp.Required(),
			),
			mcp.WithString("path",
				mcp.Description("File path relative to the repository root"),
				mcp.Required(),
			),
		),
		s.handlers.HandleReadFile,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("search_code",
			mcp.WithDescription("Search for patterns in repository code. Supports regex patterns."),
			mcp.WithString("query",
				mcp.Description("Search pattern (regex supported)"),
				mcp.Required(),
			),
			mcp.WithString("repo",
				mcp.Description("Specific repository to search, or empty for all repositories"),
			),
			mcp.WithString("file_pattern",
				mcp.Description("Glob pattern for files to search (e.g., '*.go', '*.py')"),
			),
		),
		s.handlers.HandleSearchCode,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("get_file_outline",
			mcp.WithDescription("Get the structure of a code file including functions, classes, and types."),
			mcp.WithString("repo",
				mcp.Description("Repository name"),
				mcp.Required(),
			),
			mcp.WithString("path",
				mcp.Description("File path relative to the repository root"),
				mcp.Required(),
			),
		),
		s.handlers.HandleGetFileOutline,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("find_files",
			mcp.WithDescription("Find files by name pattern across repositories."),
			mcp.WithString("pattern",
				mcp.Description("Glob pattern to match file names (e.g., '*.go', 'README*')"),
				mcp.Required(),
			),
			mcp.WithString("repo",
				mcp.Description("Specific repository to search, or empty for all repositories"),
			),
		),
		s.handlers.HandleFindFiles,
	)
}
