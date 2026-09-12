package httpserver

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/juniorosa/go-image-optimizer/backend/internal/config"
)

func NewServer(cfg config.Config, logger *slog.Logger) *http.Server {
	return &http.Server{
		Addr:              cfg.Addr(),
		Handler:           NewRouter(logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
