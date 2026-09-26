package imageprocessing

import "errors"

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
