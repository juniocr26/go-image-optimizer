package convertimage

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strconv"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageconversion"
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageprocessing"
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageresize"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/http/handler/imagehttp"
)

type useCase interface {
	Inspect(context.Context, []byte) (imageresize.Info, error)
	Execute(context.Context, []byte, imageprocessing.Format) (imageprocessing.Result, error)
}

func Inspect(logger *slog.Logger, uc useCase) http.HandlerFunc { return handle(logger, uc, true) }
func Process(logger *slog.Logger, uc useCase) http.HandlerFunc { return handle(logger, uc, false) }

func handle(logger *slog.Logger, uc useCase, inspect bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r.MultipartForm != nil {
				_ = r.MultipartForm.RemoveAll()
			}
			if recovered := recover(); recovered != nil {
				logger.Error("conversion request panicked", "panic", recovered)
				imagehttp.WriteJSONError(w, 500, "image could not be processed")
			}
		}()
		w.Header().Set("Cache-Control", "no-store")
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "multipart/form-data" {
			imagehttp.WriteJSONError(w, 400, "request must be multipart/form-data")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 50<<20)
		if err := r.ParseMultipartForm(8 << 20); err != nil {
			var limit *http.MaxBytesError
			if errors.As(err, &limit) {
				imagehttp.WriteJSONError(w, 413, "uploaded image is too large")
			} else {
				imagehttp.WriteJSONError(w, 400, "invalid multipart form")
			}
			return
		}
		if len(r.MultipartForm.File) != 1 || len(r.MultipartForm.File["image"]) != 1 || len(r.MultipartForm.Value["image"]) != 0 {
			imagehttp.WriteJSONError(w, 400, "exactly one image field is required")
			return
		}
		var options imageprocessing.Format
		if inspect && len(r.MultipartForm.Value) != 0 {
			imagehttp.WriteJSONError(w, 400, "inspection accepts only image")
			return
		}
		if !inspect {
			options, err = parseTarget(r)
			if err != nil {
				imagehttp.WriteJSONError(w, 400, "invalid targetFormat: choose a supported destination different from the source")
				return
			}
		}
		file, header, err := r.FormFile("image")
		if err != nil {
			imagehttp.WriteJSONError(w, 400, "invalid image field")
			return
		}
		defer file.Close()
		input, err := io.ReadAll(file)
		if err != nil {
			imagehttp.WriteJSONError(w, 400, "could not read image")
			return
		}
		if inspect {
			info, err := uc.Inspect(r.Context(), input)
			if err != nil {
				writeConversionError(w, logger, err)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(info)
			return
		}
		result, err := uc.Execute(r.Context(), input, options)
		if err != nil {
			if errors.Is(err, imageconversion.ErrInvalidTarget) {
				imagehttp.WriteJSONError(w, 400, "invalid targetFormat: choose a supported destination different from the source")
			} else {
				writeConversionError(w, logger, err)
			}
			return
		}
		w.Header().Set("Content-Type", result.ContentType)
		w.Header().Set("Content-Length", strconv.Itoa(len(result.Data)))
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": imagehttp.Filename(header.Filename, result.Format, result.ContentType, "converted")}))
		w.Header().Set("X-Source-Format", string(result.SourceFormat))
		w.Header().Set("X-Output-Format", string(result.Format))
		w.Header().Set("X-Original-Width", strconv.Itoa(result.Width))
		w.Header().Set("X-Original-Height", strconv.Itoa(result.Height))
		w.Header().Set("X-Image-Width", strconv.Itoa(result.Width))
		w.Header().Set("X-Image-Height", strconv.Itoa(result.Height))
		if _, err := w.Write(result.Data); err != nil {
			logger.Warn("could not write conversion response", "error", err)
		}
	}
}

func parseTarget(r *http.Request) (imageprocessing.Format, error) {
	values := r.MultipartForm.Value
	if len(values) != 1 || len(values["targetFormat"]) != 1 {
		return "", imageconversion.ErrInvalidTarget
	}
	target := imageprocessing.Format(values["targetFormat"][0])
	if !imageconversion.ValidTarget(target) {
		return "", imageconversion.ErrInvalidTarget
	}
	return target, nil
}
func writeConversionError(w http.ResponseWriter, logger *slog.Logger, err error) {
	if errors.Is(err, imageprocessing.ErrUnsupportedVariant) {
		imagehttp.WriteJSONError(w, 422, "Conversion of animations, APNG, AVIF sequences and multi-image TIFF/HEIF is not supported.")
		return
	}
	imagehttp.WriteProcessingError(w, logger, err)
}
