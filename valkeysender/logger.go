package valkeysender

import (
	"context"
	"log/slog"
	"os"
)

// NewLogger creates a new structured logger for valkeysender
func NewLogger(level slog.Level, logFilePath string) (*slog.Logger, error) {
	var handler slog.Handler
	
	// Create JSON handler for structured logging
	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Add component identifier to all log entries
			if a.Key == slog.SourceKey {
				return slog.Attr{}
			}
			return a
		},
	}
	
	// Always log to stdout for container environments
	handler = slog.NewJSONHandler(os.Stdout, opts)
	
	// If log file path is specified, also log to file
	if logFilePath != "" {
		logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		
		// Use file handler for file output
		fileHandler := slog.NewJSONHandler(logFile, opts)
		
		// Create a multi-handler that writes to both stdout and file
		handler = &multiHandler{
			handlers: []slog.Handler{handler, fileHandler},
		}
	}
	
	logger := slog.New(handler)
	
	// Add component context to all log entries
	logger = logger.With(
		slog.String("component", "valkeysender"),
		slog.String("version", "1.0.0"),
	)
	
	return logger, nil
}

// multiHandler implements slog.Handler to write to multiple destinations
type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, h := range m.handlers {
		if err := h.Handle(ctx, record); err != nil {
			return err
		}
	}
	return nil
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: newHandlers}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: newHandlers}
}