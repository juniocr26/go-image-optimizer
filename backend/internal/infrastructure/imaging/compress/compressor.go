package compress

import (
	"image"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/imaging"
)

const (
	// DefaultJPEGQuality keeps JPEG output visually close to the source while
	// still allowing meaningful size reduction for many camera/exported images.
	DefaultJPEGQuality = 82

	DefaultWebPQuality = 82
	DefaultAVIFQuality = 60
	DefaultHEIFQuality = 60
)

type Compressor struct {
	JPEGQuality            int
	WebPQuality            int
	AVIFQuality            int
	HEIFQuality            int
	MaxDecodedPixels       int
	MaxAnimatedFramePixels int
}

func NewCompressor() Compressor {
	return Compressor{
		JPEGQuality:            DefaultJPEGQuality,
		WebPQuality:            DefaultWebPQuality,
		AVIFQuality:            DefaultAVIFQuality,
		HEIFQuality:            DefaultHEIFQuality,
		MaxDecodedPixels:       imaging.DefaultMaxDecodedPixels,
		MaxAnimatedFramePixels: imaging.DefaultMaxAnimatedFramePixels,
	}
}

func (c Compressor) Compress(input []byte) (imagecompression.Result, error) {
	detected, err := imaging.DetectFormat(input)
	if err != nil {
		return imagecompression.Result{}, err
	}

	switch detected.Format {
	case imagecompression.FormatJPEG:
		return c.compressJPEG(input)
	case imagecompression.FormatPNG:
		return c.compressPNG(input)
	case imagecompression.FormatWebP:
		return c.compressWebP(input)
	case imagecompression.FormatAVIF:
		return c.compressAVIF(input)
	case imagecompression.FormatHEIF:
		return c.compressHEIF(input, detected.ContentType)
	case imagecompression.FormatGIF:
		return c.compressGIF(input)
	case imagecompression.FormatBMP:
		return c.compressBMP(input)
	case imagecompression.FormatTIFF:
		return c.compressTIFF(input)
	default:
		return imagecompression.Result{}, imagecompression.ErrUnsupportedFormat
	}
}

func staticResult(data []byte, format imagecompression.Format, contentType string, img image.Image) imagecompression.Result {
	bounds := img.Bounds()

	return imagecompression.Result{
		Data:        data,
		Format:      format,
		ContentType: contentType,
		Width:       bounds.Dx(),
		Height:      bounds.Dy(),
		Animated:    false,
		FrameCount:  1,
	}
}

func (c Compressor) jpegQuality() int {
	return clampQuality(c.JPEGQuality, DefaultJPEGQuality)
}

func (c Compressor) webpQuality() int {
	return clampQuality(c.WebPQuality, DefaultWebPQuality)
}

func (c Compressor) avifQuality() int {
	return clampQuality(c.AVIFQuality, DefaultAVIFQuality)
}

func (c Compressor) heifQuality() int {
	return clampQuality(c.HEIFQuality, DefaultHEIFQuality)
}

func (c Compressor) maxDecodedPixels() int {
	if c.MaxDecodedPixels <= 0 {
		return imaging.DefaultMaxDecodedPixels
	}

	return c.MaxDecodedPixels
}

func (c Compressor) maxAnimatedFramePixels() int {
	if c.MaxAnimatedFramePixels <= 0 {
		return imaging.DefaultMaxAnimatedFramePixels
	}

	return c.MaxAnimatedFramePixels
}

func clampQuality(value, fallback int) int {
	if value == 0 {
		return fallback
	}

	if value < 1 {
		return 1
	}

	if value > 100 {
		return 100
	}

	return value
}
