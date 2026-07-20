package handlers

import (
	"clarus/internal/app/rfc"
	"clarus/internal/app/repo"
	"clarus/internal/config"
	"clarus/internal/logging"
)

// Context provides dependencies to all handlers
type Context struct {
	Config      *config.Config
	RFCService  rfc.Service
	RepoService repo.Service
	Logger      *logging.Logger
}

// NewContext creates a new handler context with all dependencies
func NewContext(cfg *config.Config, logger *logging.Logger) (*Context, error) {
	rfcService, err := rfc.NewService(&cfg.RFCDocs)
	if err != nil {
		return nil, err
	}

	repoService := repo.NewService(&cfg.Repos)

	return &Context{
		Config:      cfg,
		RFCService:  rfcService,
		RepoService: repoService,
		Logger:      logger,
	}, nil
}
