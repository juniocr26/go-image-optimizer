package imagecompression

import (
	"context"
	"errors"
)

type Format string

const (
	FormatJPEG Format = "jpeg"
	FormatPNG  Format = "png"
	FormatWebP Format = "webp"
	FormatAVIF Format = "avif"
	FormatHEIF Format = "heif"
	FormatGIF  Format = "gif"
	FormatBMP  Format = "bmp"
	FormatTIFF Format = "tiff"
)

var (
	ErrEmptyImage         = errors.New("image content is empty")
	ErrUnsupportedFormat  = errors.New("unsupported image format")
	ErrInvalidImage       = errors.New("invalid image content")
	ErrImageTooLarge      = errors.New("image dimensions exceed the configured safety limit")
	ErrAnimationTooLarge  = errors.New("animated image exceeds the configured frame safety limit")
	ErrUnsupportedVariant = errors.New("unsupported image variant")
	ErrCodecUnavailable   = errors.New("required image codec is unavailable")
)

type Result struct {
	Data        []byte
	Format      Format
	ContentType string
	Width       int
	Height      int
	Animated    bool
	FrameCount  int
}

type Compressor interface {
	Compress(input []byte) (Result, error)
}

type UseCase struct {
	compressor Compressor
}

func NewUseCase(compressor Compressor) UseCase {
	return UseCase{
		compressor: compressor,
	}
}

func (u UseCase) Execute(ctx context.Context, input []byte) (Result, error) {
	if len(input) == 0 {
		return Result{}, ErrEmptyImage
	}

	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	result, err := u.compressor.Compress(input)
	if err != nil {
		return Result{}, err
	}

	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	return result, nil
}
