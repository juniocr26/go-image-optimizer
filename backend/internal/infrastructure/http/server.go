package httpserver

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/juniorosa/go-image-optimizer/backend/internal/config"
)

const (
	readHeaderTimeout     = 5 * time.Second
	imageRequestTimeout   = 2 * time.Minute
	idleConnectionTimeout = 60 * time.Second
)

func NewServer(cfg config.Config, logger *slog.Logger) *http.Server {
	return &http.Server{
		Addr:              cfg.Addr(),
		Handler:           NewRouter(logger),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       imageRequestTimeout,
		WriteTimeout:      imageRequestTimeout,
		IdleTimeout:       idleConnectionTimeout,
	}
}
