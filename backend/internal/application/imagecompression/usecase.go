package imagecompression

import (
	"context"
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageprocessing"
)

type Format = imageprocessing.Format
type Result = imageprocessing.Result

const (
	FormatJPEG = imageprocessing.FormatJPEG
	FormatPNG  = imageprocessing.FormatPNG
	FormatWebP = imageprocessing.FormatWebP
	FormatAVIF = imageprocessing.FormatAVIF
	FormatHEIF = imageprocessing.FormatHEIF
	FormatGIF  = imageprocessing.FormatGIF
	FormatBMP  = imageprocessing.FormatBMP
	FormatTIFF = imageprocessing.FormatTIFF
)

var (
	ErrEmptyImage         = imageprocessing.ErrEmptyImage
	ErrUnsupportedFormat  = imageprocessing.ErrUnsupportedFormat
	ErrInvalidImage       = imageprocessing.ErrInvalidImage
	ErrImageTooLarge      = imageprocessing.ErrImageTooLarge
	ErrAnimationTooLarge  = imageprocessing.ErrAnimationTooLarge
	ErrUnsupportedVariant = imageprocessing.ErrUnsupportedVariant
	ErrCodecUnavailable   = imageprocessing.ErrCodecUnavailable
)

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
