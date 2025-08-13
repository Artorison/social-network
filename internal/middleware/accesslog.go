package middleware

import (
	"log/slog"
	"net/http"
	"redditclone/pkg/logger"
	"time"

	"github.com/gorilla/mux"
)

func AccessLog(logger *logger.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Info("New request",
				slog.String("method", r.Method),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("url", r.URL.Path),
				slog.String("request_time", time.Since(start).String()),
			)
		})
	}
}
