package imagecompression

import (
	"context"
	"errors"
)

type Format string

const (
	FormatJPEG Format = "jpeg"
	FormatPNG  Format = "png"
)

var (
	ErrEmptyImage        = errors.New("image content is empty")
	ErrUnsupportedFormat = errors.New("unsupported image format")
	ErrInvalidImage      = errors.New("invalid image content")
	ErrImageTooLarge     = errors.New("image dimensions exceed the configured safety limit")
)

type Result struct {
	Data        []byte
	Format      Format
	ContentType string
	Width       int
	Height      int
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
