package compressimage

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
)

func writeCompressionError(w http.ResponseWriter, logger *slog.Logger, err error) {
	switch {
	case errors.Is(err, imagecompression.ErrEmptyImage):
		writeJSONError(w, http.StatusBadRequest, "uploaded image is empty")
	case errors.Is(err, imagecompression.ErrUnsupportedFormat):
		writeJSONError(w, http.StatusUnsupportedMediaType, "unsupported image format. Use JPEG, PNG, WebP, AVIF, HEIC, GIF, BMP, or TIFF")
	case errors.Is(err, imagecompression.ErrInvalidImage):
		writeJSONError(w, http.StatusBadRequest, "image content is invalid or corrupted")
	case errors.Is(err, imagecompression.ErrImageTooLarge):
		writeJSONError(w, http.StatusRequestEntityTooLarge, "image dimensions are too large to process safely")
	case errors.Is(err, imagecompression.ErrAnimationTooLarge):
		writeJSONError(w, http.StatusRequestEntityTooLarge, "animated image has too many canvas pixels to process safely")
	case errors.Is(err, imagecompression.ErrUnsupportedVariant):
		writeJSONError(w, http.StatusUnprocessableEntity, "this image variant is not supported")
	case errors.Is(err, imagecompression.ErrCodecUnavailable):
		logger.Warn("image codec unavailable", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "required image codec is unavailable")
	default:
		logger.Warn("image compression failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "image could not be compressed")
	}
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
