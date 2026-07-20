package mcp

import (
	"clarus/internal/logging"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Config struct {
	Version                         string `env:"VERSION" envDefault:"1.0.0"`
	ToolCapabilities                bool   `env:"TOOL_CAPABILITIES" envDefault:"true"`
	ResourceCapabilitiesSubscribe   bool   `env:"RESOURCE_CAPABILITIES_SUBSCRIBE" envDefault:"true"`
	ResourceCapabilitiesListChanged bool   `env:"RESOURCE_CAPABILITIES_LIST_CHANGED" envDefault:"true"`
}

func NewMCPServer(log *logging.Logger, cfg *Config) *server.MCPServer {

	// Create server
	s := server.NewMCPServer("clarus", cfg.Version,
		server.WithToolCapabilities(cfg.ToolCapabilities),
		server.WithResourceCapabilities(
			cfg.ResourceCapabilitiesSubscribe,
			cfg.ResourceCapabilitiesListChanged,
		),
	)

	// Add tool with builder pattern
	s.AddTool(
		mcp.NewTool("search_rfcs",
			mcp.WithDescription("Search RFC documents"),
			mcp.WithString("query", mcp.Required()),
		),
		nil,
	)

	return s
}

// ServeStdio runs the server over stdio (blocking)
func ServeStdio(s *server.MCPServer) error {
	return server.ServeStdio(s)
}
