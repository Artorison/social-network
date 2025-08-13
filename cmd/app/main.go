package main

import (
	"log/slog"
	"os"

	"github.com/Artorison/social-network/config"
	"github.com/Artorison/social-network/internal/app"
	"github.com/Artorison/social-network/internal/sessions"
	"github.com/Artorison/social-network/pkg/logger"
)

func main() {
	cfg := config.MustLoad()

	l := logger.NewSlogLogger(slog.LevelInfo, logger.TEXTLogs)

	sessions.MustSecretKey(os.Getenv("SECRET_KEY"))
	sessions.MustPepperKey(os.Getenv("PEPPER_SECRET"))

	app.Run(cfg, l)
}
