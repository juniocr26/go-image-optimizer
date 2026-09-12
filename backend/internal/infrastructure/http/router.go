package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/http/handler"
)

func NewRouter(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.Health(logger))
	mux.HandleFunc("POST /images/compress", handler.ProcessImage(logger))

	return mux
}
