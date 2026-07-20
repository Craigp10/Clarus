package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

// registerPrompts registers all MCP prompts with the server
func (s *Server) registerPrompts() {
	// Understand Domain prompt
	s.mcpServer.AddPrompt(
		mcp.NewPrompt("understand_domain",
			mcp.WithPromptDescription("Generate a comprehensive understanding of the business domain based on available RFCs and code"),
		),
		func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return &mcp.GetPromptResult{
				Description: "Understand the business domain",
				Messages: []mcp.PromptMessage{
					{
						Role: mcp.RoleUser,
						Content: mcp.TextContent{
							Type: "text",
							Text: `Please analyze the available RFC documents and repositories to provide a comprehensive overview of the business domain.

Use the following tools to gather information:
1. Call list_rfcs to see all available RFC documents
2. Call list_repos to see all available repositories
3. Read key RFCs that describe the core business concepts

Then provide:
- A summary of what this system/product does
- Key business concepts and terminology
- Main workflows and processes
- Important architectural decisions
- How the repositories relate to the business requirements`,
						},
					},
				},
			}, nil
		},
	)

	// Explain Feature prompt
	s.mcpServer.AddPrompt(
		mcp.NewPrompt("explain_feature",
			mcp.WithPromptDescription("Explain how a specific feature works based on RFCs and code"),
			mcp.WithArgument("feature_name",
				mcp.ArgumentDescription("Name or description of the feature to explain"),
				mcp.RequiredArgument(),
			),
		),
		func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			featureName := ""
			if fn, ok := request.Params.Arguments["feature_name"].(string); ok {
				featureName = fn
			}

			return &mcp.GetPromptResult{
				Description: "Explain feature: " + featureName,
				Messages: []mcp.PromptMessage{
					{
						Role: mcp.RoleUser,
						Content: mcp.TextContent{
							Type: "text",
							Text: `Please explain how the "` + featureName + `" feature works.

Use the available tools to:
1. Search RFCs for documentation about this feature using search_rfcs
2. Find related code using search_code
3. Read relevant files to understand the implementation

Provide:
- What this feature does from a business perspective
- How it's documented in RFCs
- Where the implementation lives in the codebase
- Key files and functions involved
- Any important configuration or dependencies`,
						},
					},
				},
			}, nil
		},
	)

	// Plan Implementation prompt
	s.mcpServer.AddPrompt(
		mcp.NewPrompt("plan_implementation",
			mcp.WithPromptDescription("Create an implementation plan for a new feature based on existing patterns"),
			mcp.WithArgument("feature_description",
				mcp.ArgumentDescription("Description of the feature to implement"),
				mcp.RequiredArgument(),
			),
			mcp.WithArgument("constraints",
				mcp.ArgumentDescription("Any constraints or requirements to consider"),
			),
		),
		func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			featureDesc := ""
			if fd, ok := request.Params.Arguments["feature_description"].(string); ok {
				featureDesc = fd
			}

			constraints := ""
			if c, ok := request.Params.Arguments["constraints"].(string); ok {
				constraints = c
			}

			constraintText := ""
			if constraints != "" {
				constraintText = "\n\nConstraints to consider: " + constraints
			}

			return &mcp.GetPromptResult{
				Description: "Plan implementation for: " + featureDesc,
				Messages: []mcp.PromptMessage{
					{
						Role: mcp.RoleUser,
						Content: mcp.TextContent{
							Type: "text",
							Text: `Please create an implementation plan for the following feature:

` + featureDesc + constraintText + `

Use the available tools to:
1. Search for similar features or patterns in existing RFCs
2. Find related code implementations using search_code
3. Analyze the impact on existing systems using analyze_impact
4. Explore repository structures to understand where new code should go

Provide:
- Summary of the feature requirements
- Relevant existing RFCs and how they inform the implementation
- Similar patterns found in the codebase
- Suggested implementation approach
- Files that need to be created or modified
- Potential risks or considerations
- Testing strategy`,
						},
					},
				},
			}, nil
		},
	)

	// Review Against RFCs prompt
	s.mcpServer.AddPrompt(
		mcp.NewPrompt("review_against_rfcs",
			mcp.WithPromptDescription("Review proposed changes against RFC requirements and specifications"),
			mcp.WithArgument("change_description",
				mcp.ArgumentDescription("Description of the changes to review"),
				mcp.RequiredArgument(),
			),
			mcp.WithArgument("rfc_paths",
				mcp.ArgumentDescription("Specific RFC paths to check against (comma-separated)"),
			),
		),
		func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			changeDesc := ""
			if cd, ok := request.Params.Arguments["change_description"].(string); ok {
				changeDesc = cd
			}

			rfcPaths := ""
			if rp, ok := request.Params.Arguments["rfc_paths"].(string); ok {
				rfcPaths = rp
			}

			rfcText := ""
			if rfcPaths != "" {
				rfcText = "\n\nSpecific RFCs to review against: " + rfcPaths
			}

			return &mcp.GetPromptResult{
				Description: "Review changes against RFCs",
				Messages: []mcp.PromptMessage{
					{
						Role: mcp.RoleUser,
						Content: mcp.TextContent{
							Type: "text",
							Text: `Please review the following proposed changes against RFC requirements:

` + changeDesc + rfcText + `

Use the available tools to:
1. Find relevant RFCs using search_rfcs
2. Read the full RFC content using read_rfc
3. Extract requirements sections using get_rfc_structure

Provide:
- List of relevant RFCs and their requirements
- How the changes align with or deviate from specifications
- Any requirements that might be violated
- Missing considerations from the RFCs
- Recommendations for compliance`,
						},
					},
				},
			}, nil
		},
	)

	// Trace Workflow prompt
	s.mcpServer.AddPrompt(
		mcp.NewPrompt("trace_workflow",
			mcp.WithPromptDescription("Trace a business workflow through RFCs and implementation code"),
			mcp.WithArgument("workflow_name",
				mcp.ArgumentDescription("Name of the workflow to trace"),
				mcp.RequiredArgument(),
			),
			mcp.WithArgument("detail_level",
				mcp.ArgumentDescription("Level of detail: 'high', 'medium', or 'detailed'"),
			),
		),
		func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			workflowName := ""
			if wn, ok := request.Params.Arguments["workflow_name"].(string); ok {
				workflowName = wn
			}

			detailLevel := "medium"
			if dl, ok := request.Params.Arguments["detail_level"].(string); ok {
				detailLevel = dl
			}

			return &mcp.GetPromptResult{
				Description: "Trace workflow: " + workflowName,
				Messages: []mcp.PromptMessage{
					{
						Role: mcp.RoleUser,
						Content: mcp.TextContent{
							Type: "text",
							Text: `Please trace the "` + workflowName + `" workflow through both documentation and code.

Detail level: ` + detailLevel + `

Use the available tools to:
1. Find workflow documentation using explain_workflow and search_rfcs
2. Find implementation code using search_code
3. Read relevant files to understand the flow

Provide:
- Overview of the workflow purpose
- Step-by-step breakdown from RFCs
- Corresponding code implementations for each step
- Entry points and exit points
- Data transformations along the way
- Error handling and edge cases
- Dependencies and external systems involved`,
						},
					},
				},
			}, nil
		},
	)
}
