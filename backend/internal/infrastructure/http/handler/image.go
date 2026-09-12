package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
)

const (
	maxImageUploadBytes = 25 << 20
	maxMultipartMemory  = 8 << 20
)

func ProcessImage(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		contentType := fileHeader.Header.Get("Content-Type")
		if contentType == "" || contentType == "application/octet-stream" {
			contentType = http.DetectContentType(imageBytes)
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", strconv.Itoa(len(imageBytes)))
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
			"filename": cleanFilename(fileHeader.Filename),
		}))
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write(imageBytes); err != nil {
			logger.Warn("failed to write image response", "error", err)
		}
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

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
