package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
)

const (
	maxImageUploadBytes = 50 << 20
	maxMultipartMemory  = 8 << 20
)

type compressImageUseCase interface {
	Execute(ctx context.Context, input []byte) (imagecompression.Result, error)
}

func ProcessImage(logger *slog.Logger, useCase compressImageUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("image compression request panicked", "panic", recovered)
				writeJSONError(w, http.StatusInternalServerError, "image could not be compressed")
			}
		}()

		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "multipart/form-data" {
			writeJSONError(w, http.StatusBadRequest, "request must be multipart/form-data")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxImageUploadBytes)

		if err := r.ParseMultipartForm(maxMultipartMemory); err != nil {
			var maxBytesError *http.MaxBytesError
			if errors.As(err, &maxBytesError) {
				writeJSONError(w, http.StatusRequestEntityTooLarge, "uploaded image is too large")
				return
			}

			writeJSONError(w, http.StatusBadRequest, "invalid multipart form")
			return
		}

		if r.MultipartForm != nil {
			defer r.MultipartForm.RemoveAll()
		}

		file, fileHeader, err := r.FormFile("image")
		if err != nil {
			if errors.Is(err, http.ErrMissingFile) {
				writeJSONError(w, http.StatusBadRequest, "image field is required")
				return
			}

			logger.Warn("failed to read image field", "error", err)
			writeJSONError(w, http.StatusBadRequest, "invalid image field")
			return
		}
		defer file.Close()

		imageBytes, err := io.ReadAll(file)
		if err != nil {
			var maxBytesError *http.MaxBytesError
			if errors.As(err, &maxBytesError) {
				writeJSONError(w, http.StatusRequestEntityTooLarge, "uploaded image is too large")
				return
			}

			logger.Warn("failed to read uploaded image", "error", err)
			writeJSONError(w, http.StatusBadRequest, "failed to read uploaded image")
			return
		}

		result, err := useCase.Execute(r.Context(), imageBytes)
		if err != nil {
			writeCompressionError(w, logger, err)
			return
		}

		w.Header().Set("Content-Type", result.ContentType)
		w.Header().Set("Content-Length", strconv.Itoa(len(result.Data)))
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
			"filename": compressedFilename(fileHeader.Filename, result.Format, result.ContentType),
		}))
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write(result.Data); err != nil {
			logger.Warn("failed to write image response", "error", err)
		}
	}
}

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

func cleanFilename(filename string) string {
	filename = strings.ReplaceAll(filename, "\\", "/")
	filename = path.Base(filename)

	if filename == "." || filename == "/" || filename == "" {
		return "image"
	}

	return filename
}

func compressedFilename(filename string, format imagecompression.Format, contentType string) string {
	filename = cleanFilename(filename)
	extension := strings.ToLower(path.Ext(filename))
	expectedExtension := defaultExtension(format, contentType)

	if !extensionMatchesFormat(extension, format) {
		extension = expectedExtension
	}

	base := strings.TrimSuffix(filename, path.Ext(filename))
	if base == "" || base == "." || base == "/" {
		base = "image"
	}

	return base + "_compressed" + extension
}

func extensionMatchesFormat(extension string, format imagecompression.Format) bool {
	switch format {
	case imagecompression.FormatJPEG:
		return extension == ".jpg" || extension == ".jpeg"
	case imagecompression.FormatPNG:
		return extension == ".png"
	case imagecompression.FormatWebP:
		return extension == ".webp"
	case imagecompression.FormatAVIF:
		return extension == ".avif"
	case imagecompression.FormatHEIF:
		return extension == ".heic" || extension == ".heif"
	case imagecompression.FormatGIF:
		return extension == ".gif"
	case imagecompression.FormatBMP:
		return extension == ".bmp"
	case imagecompression.FormatTIFF:
		return extension == ".tif" || extension == ".tiff"
	default:
		return false
	}
}

func defaultExtension(format imagecompression.Format, contentType string) string {
	switch format {
	case imagecompression.FormatJPEG:
		return ".jpg"
	case imagecompression.FormatPNG:
		return ".png"
	case imagecompression.FormatWebP:
		return ".webp"
	case imagecompression.FormatAVIF:
		return ".avif"
	case imagecompression.FormatHEIF:
		if contentType == "image/heif" {
			return ".heif"
		}
		return ".heic"
	case imagecompression.FormatGIF:
		return ".gif"
	case imagecompression.FormatBMP:
		return ".bmp"
	case imagecompression.FormatTIFF:
		return ".tiff"
	default:
		return ".jpg"
	}
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
