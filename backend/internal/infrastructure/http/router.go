package httpserver

import (
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageconversion"
	convertimage "github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/http/handler/convert_image"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/imaging/convert"
	"log/slog"
	"net/http"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageresize"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/http/handler"
	compressimage "github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/http/handler/compress_image"
	resizeimage "github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/http/handler/resize_image"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/imaging/compress"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/imaging/resize"
)

func NewRouter(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	compressImage := imagecompression.NewUseCase(compress.NewCompressor())

	mux.HandleFunc("GET /health", handler.Health(logger))
	mux.HandleFunc("POST /images/compress", compressimage.ProcessImage(logger, compressImage))

	resizeImage := imageresize.NewUseCase(resize.Decoder{})
	mux.HandleFunc("POST /images/resize/info", resizeimage.Inspect(logger, resizeImage))
	mux.HandleFunc("POST /images/resize", resizeimage.Process(logger, resizeImage))

	conversion := imageconversion.NewUseCase(convert.Processor{})
	mux.HandleFunc("POST /images/convert", convertimage.Process(logger, conversion))
	mux.HandleFunc("POST /images/convert/info", convertimage.Inspect(logger, conversion))
	return mux
}
