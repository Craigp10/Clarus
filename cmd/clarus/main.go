package main

import (
	"log"

	"clarus/internal/logging"
	"clarus/internal/mcp"

	"github.com/caarlos0/env/v8"
	"github.com/joho/godotenv"
)

type Config struct {
	Logging logging.Config `envPrefix:"LOG_"`
	MCP     mcp.Config     `envPrefix:"MCP_"`
}

func main() {
	// Load .env file (optional - won't error if missing)
	_ = godotenv.Load()

	// Parse environment variables into config
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	// Create logger
	logger := logging.NewLogger(cfg.Logging)

	// Create MCP server
	s := mcp.NewMCPServer(logger, &cfg.MCP)

	// Run server (blocks until stdin closes)
	if err := mcp.ServeStdio(s); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
