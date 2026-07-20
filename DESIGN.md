# Business Context Agent MCP Server - Implementation Plan

## Overview

Build an MCP server in Go that acts as a business context agent, providing access to RFC documents and multiple repositories to answer business logic questions, explain workflows, and help plan new work.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                   Business Context Agent                     │
│                       (MCP Server)                           │
├─────────────────────────────────────────────────────────────┤
│  cmd/clarus/main.go          │  Entry point                 │
├──────────────────────────────┼──────────────────────────────┤
│  internal/mcp/               │  MCP server, tools,          │
│    server.go, tools.go       │  resources, prompts          │
│    resources.go, prompts.go  │                              │
│    handlers/                 │  Tool/resource handlers      │
├──────────────────────────────┼──────────────────────────────┤
│  internal/app/rfc/           │  RFC parsing, indexing,      │
│    service.go, parser.go     │  search                      │
├──────────────────────────────┼──────────────────────────────┤
│  internal/app/repo/          │  Repo exploration,           │
│    service.go, walker.go     │  code search                 │
├──────────────────────────────┼──────────────────────────────┤
│  internal/config/            │  Env-based configuration     │
│  internal/utils/             │  File utilities, markdown    │
│  internal/logging/           │  Structured logging          │
└─────────────────────────────────────────────────────────────┘
              │ stdio transport
              ▼
       MCP Client (Claude)
```

## Configuration

Environment variables:
- `RFC_DOCS_ROOT` (required) - Path to RFC markdown documents
- `REPOS_ROOT` (required) - Path to directory containing repositories
- `CLARUS_LOG_LEVEL` (optional, default: "info")
- `CLARUS_MAX_FILE_SIZE` (optional, default: 10MB)
- `CLARUS_EXCLUDE_PATTERNS` (optional, default: ".git,node_modules,vendor")

## MCP Tools (14 total)

### RFC Tools
| Tool | Description |
|------|-------------|
| `search_rfcs` | Full-text search across all RFC documents |
| `read_rfc` | Read complete content of a specific RFC |
| `list_rfcs` | List all RFCs with titles and summaries |
| `get_rfc_structure` | Extract headings/sections of an RFC |
| `search_rfc_sections` | Search within specific RFC sections |

### Repository Tools
| Tool | Description |
|------|-------------|
| `list_repos` | List all repositories under REPOS_ROOT |
| `explore_repo_structure` | Get directory tree of a repository |
| `read_file` | Read a specific file from a repository |
| `search_code` | Search for patterns in repository code |
| `get_file_outline` | Get structure of a code file (functions, types) |
| `find_files` | Find files by name pattern |

### Business Logic & Planning Tools
| Tool | Description |
|------|-------------|
| `explain_workflow` | Extract and explain a workflow from RFCs |
| `find_related_rfcs` | Find RFCs related to a topic |
| `analyze_impact` | Analyze which RFCs/repos affected by a change |

## MCP Resources

Static:
- `rfc://index` - Complete RFC index
- `repo://index` - Repository index

Dynamic (templates):
- `rfc://{path}` - RFC document content
- `repo://{name}/tree` - Repository structure
- `repo://{name}/file/{path}` - File content

## MCP Prompts

| Prompt | Purpose |
|--------|---------|
| `understand_domain` | Comprehensive domain overview |
| `explain_feature` | Explain how a feature works |
| `plan_implementation` | Create implementation plan for new feature |
| `review_against_rfcs` | Review changes against RFC requirements |
| `trace_workflow` | Trace workflow through RFCs and code |

## Files to Create

```
cmd/clarus/main.go                      # Entry point
internal/config/config.go               # Config struct
internal/config/env.go                  # Env loading
internal/config/validation.go           # Config validation
internal/logging/logger.go              # Structured logging
internal/mcp/server.go                  # MCP server setup
internal/mcp/tools.go                   # Tool registrations
internal/mcp/resources.go               # Resource registrations
internal/mcp/prompts.go                 # Prompt registrations
internal/mcp/handlers/context.go        # Handler dependencies
internal/mcp/handlers/rfc_handlers.go   # RFC tool handlers
internal/mcp/handlers/repo_handlers.go  # Repo tool handlers
internal/mcp/handlers/resource_handlers.go
internal/mcp/handlers/prompt_handlers.go
internal/app/rfc/models.go              # RFC domain models
internal/app/rfc/service.go             # RFC service interface
internal/app/rfc/parser.go              # Markdown parsing
internal/app/rfc/indexer.go             # RFC indexing
internal/app/repo/models.go             # Repo domain models
internal/app/repo/service.go            # Repo service interface
internal/app/repo/walker.go             # Directory walking
internal/utils/files.go                 # File utilities
internal/utils/sanitize.go              # Path security
internal/utils/markdown/parser.go       # Goldmark wrapper
```

## Dependencies

```go
require (
    github.com/mark3labs/mcp-go v0.27.0
    github.com/yuin/goldmark v1.7.0
)
```

## Implementation Order

1. **Phase 1: Foundation**
   - config package (config.go, env.go, validation.go)
   - logging package
   - utils/files.go, utils/sanitize.go

2. **Phase 2: Core Services**
   - app/rfc/models.go, service.go, parser.go
   - app/repo/models.go, service.go, walker.go

3. **Phase 3: MCP Server**
   - mcp/server.go
   - mcp/handlers/context.go
   - mcp/tools.go + handlers
   - mcp/resources.go + handlers
   - mcp/prompts.go + handlers

4. **Phase 4: Entry Point**
   - cmd/clarus/main.go

## Verification

1. **Build**: `go build -o clarus ./cmd/clarus`

2. **Test configuration**:
   ```bash
   RFC_DOCS_ROOT=/path/to/rfcs REPOS_ROOT=/path/to/repos ./clarus
   ```

3. **Test with MCP Inspector** (if available):
   ```bash
   npx @modelcontextprotocol/inspector ./clarus
   ```

4. **Add to Claude Code config** (`~/.claude/claude_desktop_config.json`):
   ```json
   {
     "mcpServers": {
       "business-context": {
         "command": "/path/to/clarus",
         "env": {
           "RFC_DOCS_ROOT": "/path/to/rfcs",
           "REPOS_ROOT": "/path/to/repos"
         }
       }
     }
   }
   ```
