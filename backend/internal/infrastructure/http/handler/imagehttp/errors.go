package imagehttp

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageprocessing"
)

func WriteProcessingError(w http.ResponseWriter, logger *slog.Logger, err error) {
	switch {
	case errors.Is(err, imageprocessing.ErrEmptyImage):
		WriteJSONError(w, http.StatusBadRequest, "uploaded image is empty")
	case errors.Is(err, imageprocessing.ErrUnsupportedFormat):
		WriteJSONError(w, http.StatusUnsupportedMediaType, "unsupported image format. Use JPEG, PNG, WebP, AVIF, HEIC, GIF, BMP, or TIFF")
	case errors.Is(err, imageprocessing.ErrInvalidImage):
		WriteJSONError(w, http.StatusBadRequest, "image content is invalid or corrupted")
	case errors.Is(err, imageprocessing.ErrImageTooLarge):
		WriteJSONError(w, http.StatusRequestEntityTooLarge, "image dimensions are too large to process safely")
	case errors.Is(err, imageprocessing.ErrAnimationTooLarge):
		WriteJSONError(w, http.StatusRequestEntityTooLarge, "animated image has too many canvas pixels to process safely")
	case errors.Is(err, imageprocessing.ErrUnsupportedVariant):
		WriteJSONError(w, http.StatusUnprocessableEntity, "this image variant is not supported")
	case errors.Is(err, imageprocessing.ErrCodecUnavailable):
		logger.Warn("image codec unavailable", "error", err)
		WriteJSONError(w, http.StatusInternalServerError, "required image codec is unavailable")
	default:
		logger.Warn("image processing failed", "error", err)
		WriteJSONError(w, http.StatusInternalServerError, "image could not be processed")
	}
}

func WriteJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
