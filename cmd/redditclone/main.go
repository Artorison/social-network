package main

import (
	"log/slog"
	"os"
	"redditclone/config"
	"redditclone/internal/app"
	"redditclone/internal/sessions"
	"redditclone/pkg/logger"
)

func main() {

	cfg := config.MustLoad()

	l := logger.NewSlogLogger(slog.LevelInfo, logger.TEXTLogs)

	sessions.MustSecretKey(os.Getenv("SECRET_KEY_REDDIT"))
	sessions.MustPepperKey(os.Getenv("PEPPER_SECRET"))

	app.Run(cfg, l)
}
