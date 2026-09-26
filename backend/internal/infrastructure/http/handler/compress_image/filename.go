package compressimage

import (
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/http/handler/imagehttp"
)

func compressedFilename(filename string, format imagecompression.Format, contentType string) string {
	return imagehttp.Filename(filename, format, contentType, "compressed")
}
