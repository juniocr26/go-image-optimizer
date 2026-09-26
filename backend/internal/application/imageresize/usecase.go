package imageresize

import (
	"context"
	"errors"
	"math"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageprocessing"
)

var ErrInvalidOptions = errors.New("invalid resize options")

const MaxPixels = 32_000_000
const MaxFramePixels = 64_000_000

type Info struct {
	Width       int                    `json:"width"`
	Height      int                    `json:"height"`
	Format      imageprocessing.Format `json:"format"`
	ContentType string                 `json:"contentType"`
	FrameCount  int                    `json:"frameCount"`
}

type Options struct {
	Mode            string
	Width           int
	Height          int
	Reduction       int
	KeepAspectRatio bool
	// Axis is the last dimension edited by the user when the ratio is locked.
	Axis string
}

type Result struct {
	imageprocessing.Result
	OriginalWidth  int
	OriginalHeight int
}

// A decoded source lives only for this request. The application decides the
// target dimensions; the imaging implementation owns pixels and codecs.
type Source interface {
	Info() Info
	Resize(ctx context.Context, width, height int) (imageprocessing.Result, error)
}

type Decoder interface {
	Decode(input []byte) (Source, error)
}

type UseCase struct{ decoder Decoder }

func NewUseCase(decoder Decoder) UseCase { return UseCase{decoder: decoder} }

func (u UseCase) Inspect(ctx context.Context, input []byte) (Info, error) {
	source, err := u.decode(ctx, input)
	if err != nil {
		return Info{}, err
	}
	return source.Info(), nil
}

func (u UseCase) Execute(ctx context.Context, input []byte, options Options) (Result, error) {
	if err := options.Validate(); err != nil {
		return Result{}, err
	}
	source, err := u.decode(ctx, input)
	if err != nil {
		return Result{}, err
	}
	info := source.Info()
	w, h, err := Target(info, options)
	if err != nil {
		return Result{}, err
	}
	var output imageprocessing.Result
	if w == info.Width && h == info.Height {
		output = imageprocessing.Result{Data: input, Format: info.Format, ContentType: info.ContentType, Width: w, Height: h, FrameCount: info.FrameCount, Animated: info.FrameCount > 1}
	} else {
		output, err = source.Resize(ctx, w, h)
		if err != nil {
			return Result{}, err
		}
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	return Result{Result: output, OriginalWidth: info.Width, OriginalHeight: info.Height}, nil
}

func (u UseCase) decode(ctx context.Context, input []byte) (Source, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(input) == 0 {
		return nil, imageprocessing.ErrEmptyImage
	}
	source, err := u.decoder.Decode(input)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return source, nil
}

func (o Options) Validate() error {
	switch o.Mode {
	case "pixels":
		if o.Width < 1 || o.Height < 1 || o.Width > MaxPixels || o.Height > MaxPixels {
			return ErrInvalidOptions
		}
		if o.Axis != "width" && o.Axis != "height" {
			return ErrInvalidOptions
		}
	case "percentage":
		if o.Reduction != 25 && o.Reduction != 50 && o.Reduction != 75 {
			return ErrInvalidOptions
		}
	default:
		return ErrInvalidOptions
	}
	return nil
}

func Target(info Info, o Options) (int, int, error) {
	if err := o.Validate(); err != nil {
		return 0, 0, err
	}
	if info.Width < 1 || info.Height < 1 || info.FrameCount < 1 {
		return 0, 0, imageprocessing.ErrInvalidImage
	}
	w, h := o.Width, o.Height
	if o.Mode == "percentage" || o.KeepAspectRatio {
		scale := float64(100-o.Reduction) / 100
		if o.Mode == "pixels" {
			scale = float64(o.Width) / float64(info.Width)
			if o.Axis == "height" {
				scale = float64(o.Height) / float64(info.Height)
			}
		}
		// Check float bounds before converting to int, including hostile ratios.
		wf, hf := math.Round(float64(info.Width)*scale), math.Round(float64(info.Height)*scale)
		if wf > MaxPixels || hf > MaxPixels {
			return 0, 0, imageprocessing.ErrImageTooLarge
		}
		w, h = max(1, int(wf)), max(1, int(hf))
	}
	if w > MaxPixels/h {
		return 0, 0, imageprocessing.ErrImageTooLarge
	}
	if info.FrameCount > 1 && w > MaxFramePixels/h/info.FrameCount {
		return 0, 0, imageprocessing.ErrAnimationTooLarge
	}
	return w, h, nil
}
