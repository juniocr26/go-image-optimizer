package resizeimage

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strconv"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageprocessing"
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageresize"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/http/handler/imagehttp"
)

type useCase interface {
	Inspect(context.Context, []byte) (imageresize.Info, error)
	Execute(context.Context, []byte, imageresize.Options) (imageresize.Result, error)
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
				logger.Error("resize request panicked", "panic", recovered)
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
		var options imageresize.Options
		if !inspect {
			options, err = parseOptions(r)
			if err != nil {
				imagehttp.WriteJSONError(w, 400, "invalid resize options")
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
				writeResizeError(w, logger, err)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(info)
			return
		}
		result, err := uc.Execute(r.Context(), input, options)
		if err != nil {
			if errors.Is(err, imageresize.ErrInvalidOptions) {
				imagehttp.WriteJSONError(w, 400, "invalid resize options")
			} else {
				writeResizeError(w, logger, err)
			}
			return
		}
		w.Header().Set("Content-Type", result.ContentType)
		w.Header().Set("Content-Length", strconv.Itoa(len(result.Data)))
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": imagehttp.Filename(header.Filename, result.Format, result.ContentType, "resized")}))
		w.Header().Set("X-Original-Width", strconv.Itoa(result.OriginalWidth))
		w.Header().Set("X-Original-Height", strconv.Itoa(result.OriginalHeight))
		w.Header().Set("X-Image-Width", strconv.Itoa(result.Width))
		w.Header().Set("X-Image-Height", strconv.Itoa(result.Height))
		if _, err := w.Write(result.Data); err != nil {
			logger.Warn("could not write resize response", "error", err)
		}
	}
}

func parseOptions(r *http.Request) (imageresize.Options, error) {
	values := r.MultipartForm.Value
	for _, v := range values {
		if len(v) != 1 {
			return imageresize.Options{}, imageresize.ErrInvalidOptions
		}
	}
	value := func(key string) string {
		if len(values[key]) == 1 {
			return values[key][0]
		}
		return ""
	}
	o := imageresize.Options{Mode: value("mode"), Axis: value("axis"), KeepAspectRatio: true, WithoutEnlargement: true}
	if o.Axis == "" {
		o.Axis = "width"
	}
	for key, destination := range map[string]*bool{"keepAspectRatio": &o.KeepAspectRatio, "withoutEnlargement": &o.WithoutEnlargement} {
		if v := value(key); len(values[key]) != 0 {
			if v != "true" && v != "false" {
				return o, imageresize.ErrInvalidOptions
			}
			*destination = v == "true"
		}
	}
	var err error
	if o.Mode == "pixels" {
		o.Width, err = strconv.Atoi(value("width"))
		if err != nil {
			return o, imageresize.ErrInvalidOptions
		}
		o.Height, err = strconv.Atoi(value("height"))
		if err != nil {
			return o, imageresize.ErrInvalidOptions
		}
	} else if o.Mode == "percentage" {
		o.Reduction, err = strconv.Atoi(value("reduction"))
		if err != nil {
			return o, imageresize.ErrInvalidOptions
		}
	}
	return o, o.Validate()
}

func writeResizeError(w http.ResponseWriter, logger *slog.Logger, err error) {
	if errors.Is(err, imageprocessing.ErrUnsupportedVariant) {
		imagehttp.WriteJSONError(w, http.StatusUnprocessableEntity, "This image variant cannot be resized. Animated AVIF/APNG and multi-page TIFF/HEIF are not supported.")
		return
	}
	imagehttp.WriteProcessingError(w, logger, err)
}
