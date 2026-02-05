package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/caarlos0/env/v8"
)

type Config struct {
	Level     string    `env:"LOG_LEVEL" envDefault:"info"`
	Format    string    `env:"LOG_FORMAT" envDefault:"text"`
	AddSource bool      `env:"LOG_ADD_SOURCE" envDefault:"false"`
	Output    io.Writer `env:"-"`
}

type Logger struct {
	*slog.Logger
}

// ConfigFromEnv creates a Config from environment variables
func ConfigFromEnv() (Config, error) {
	cfg := Config{Output: os.Stdout}
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func NewLogger(cfg Config) *Logger {
	level := parseLevel(cfg.Level)

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	output := cfg.Output
	if output == nil {
		output = os.Stdout
	}

	var handler slog.Handler
	switch strings.ToLower(cfg.Format) {
	case "json":
		handler = slog.NewJSONHandler(output, opts)
	default:
		handler = slog.NewTextHandler(output, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

