package convert

import (
	"bytes"
	"context"
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageprocessing"
	"golang.org/x/image/draw"
	"image"
	"image/jpeg"
	"image/png"
	"math"
)

// Preview uses the processing decoder and its source/variant/resource limits.
// Animations remain explicit unsupported previews rather than silently flattened.
func (Processor) Preview(ctx context.Context, input []byte) (imageprocessing.Result, error) {
	if len(input) > 50<<20 {
		return imageprocessing.Result{}, imageprocessing.ErrImageTooLarge
	}
	info, img, err := decode(ctx, input)
	if err != nil {
		return imageprocessing.Result{}, err
	}
	if info.FrameCount != 1 || img == nil {
		return imageprocessing.Result{}, imageprocessing.ErrUnsupportedVariant
	}
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	ratio := math.Min(1, math.Min(1200/float64(w), 1200/float64(h)))
	w, h = max(1, int(math.Round(float64(w)*ratio))), max(1, int(math.Round(float64(h)*ratio)))
	thumb := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.CatmullRom.Scale(thumb, thumb.Bounds(), img, img.Bounds(), draw.Src, nil)
	if err := ctx.Err(); err != nil {
		return imageprocessing.Result{}, err
	}
	var out bytes.Buffer
	format, mime := imageprocessing.FormatJPEG, "image/jpeg"
	if thumb.Opaque() {
		err = jpeg.Encode(&out, thumb, &jpeg.Options{Quality: 82})
	} else {
		format, mime = imageprocessing.FormatPNG, "image/png"
		err = png.Encode(&out, thumb)
	}
	if err != nil {
		return imageprocessing.Result{}, err
	}
	if err := ctx.Err(); err != nil {
		return imageprocessing.Result{}, err
	}
	return imageprocessing.Result{Data: out.Bytes(), Format: format, ContentType: mime, Width: w, Height: h, SourceFormat: info.Format, FrameCount: 1}, nil
}
