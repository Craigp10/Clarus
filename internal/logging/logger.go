package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Level represents a log level
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel parses a string into a Level
func ParseLevel(s string) Level {
	switch s {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// Logger provides structured logging
type Logger struct {
	mu       sync.Mutex
	level    Level
	format   string
	output   io.Writer
	fields   map[string]any
}

// New creates a new Logger
func New(level, format string) *Logger {
	return &Logger{
		level:  ParseLevel(level),
		format: format,
		output: os.Stderr, // MCP servers should log to stderr, not stdout
		fields: make(map[string]any),
	}
}

// WithField returns a new logger with an additional field
func (l *Logger) WithField(key string, value any) *Logger {
	newFields := make(map[string]any, len(l.fields)+1)
	for k, v := range l.fields {
		newFields[k] = v
	}
	newFields[key] = value

	return &Logger{
		level:  l.level,
		format: l.format,
		output: l.output,
		fields: newFields,
	}
}

// WithFields returns a new logger with additional fields
func (l *Logger) WithFields(fields map[string]any) *Logger {
	newFields := make(map[string]any, len(l.fields)+len(fields))
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &Logger{
		level:  l.level,
		format: l.format,
		output: l.output,
		fields: newFields,
	}
}

// Debug logs at debug level
func (l *Logger) Debug(msg string, keyvals ...any) {
	l.log(LevelDebug, msg, keyvals...)
}

// Info logs at info level
func (l *Logger) Info(msg string, keyvals ...any) {
	l.log(LevelInfo, msg, keyvals...)
}

// Warn logs at warn level
func (l *Logger) Warn(msg string, keyvals ...any) {
	l.log(LevelWarn, msg, keyvals...)
}

// Error logs at error level
func (l *Logger) Error(msg string, keyvals ...any) {
	l.log(LevelError, msg, keyvals...)
}

func (l *Logger) log(level Level, msg string, keyvals ...any) {
	if level < l.level {
		return
	}

	// Merge keyvals into fields
	fields := make(map[string]any, len(l.fields)+len(keyvals)/2+2)
	for k, v := range l.fields {
		fields[k] = v
	}

	for i := 0; i < len(keyvals)-1; i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			key = fmt.Sprintf("%v", keyvals[i])
		}
		fields[key] = keyvals[i+1]
	}

	fields["time"] = time.Now().UTC().Format(time.RFC3339)
	fields["level"] = level.String()
	fields["msg"] = msg

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.format == "json" {
		l.writeJSON(fields)
	} else {
		l.writeText(level, msg, fields)
	}
}

func (l *Logger) writeJSON(fields map[string]any) {
	data, err := json.Marshal(fields)
	if err != nil {
		fmt.Fprintf(l.output, "ERROR: failed to marshal log: %v\n", err)
		return
	}
	fmt.Fprintln(l.output, string(data))
}

func (l *Logger) writeText(level Level, msg string, fields map[string]any) {
	timestamp := fields["time"]
	delete(fields, "time")
	delete(fields, "level")
	delete(fields, "msg")

	fmt.Fprintf(l.output, "%s [%s] %s", timestamp, level.String(), msg)

	for k, v := range fields {
		fmt.Fprintf(l.output, " %s=%v", k, v)
	}
	fmt.Fprintln(l.output)
}
