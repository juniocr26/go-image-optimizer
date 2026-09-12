package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/http/handler"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/imaging"
)

func NewRouter(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	compressImage := imagecompression.NewUseCase(imaging.NewCompressor())

	mux.HandleFunc("GET /health", handler.Health(logger))
	mux.HandleFunc("POST /images/compress", handler.ProcessImage(logger, compressImage))

	return mux
}
