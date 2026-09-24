package compressimage

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strconv"

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
