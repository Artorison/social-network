package logger

import (
	"log/slog"
	"os"
	"time"
)

type LogsType string

type Logger struct {
	*slog.Logger
}

const (
	JSONLogs LogsType = "json"
	TEXTLogs LogsType = "text"
)

func NewSlogLogger(level slog.Level, logsType LogsType) *Logger {
	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					return slog.String(slog.TimeKey, t.Format("2006-01-02 15:04:05"))
				}
			}
			return a
		},
	}

	var handler slog.Handler
	switch logsType {
	case JSONLogs:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	logger.Info("logger started")
	return &Logger{logger}
}
