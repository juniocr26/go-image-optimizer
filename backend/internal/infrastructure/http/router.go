package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/http/handler"
	compressimage "github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/http/handler/compress_image"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/imaging/compress"
)

func NewRouter(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	compressImage := imagecompression.NewUseCase(compress.NewCompressor())

	mux.HandleFunc("GET /health", handler.Health(logger))
	mux.HandleFunc("POST /images/compress", compressimage.ProcessImage(logger, compressImage))

	return mux
}
