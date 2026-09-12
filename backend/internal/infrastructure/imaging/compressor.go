package imaging

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
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
		MaxDecodedPixels:       DefaultMaxDecodedPixels,
		MaxAnimatedFramePixels: DefaultMaxAnimatedFramePixels,
	}
}

func (c Compressor) Compress(input []byte) (imagecompression.Result, error) {
	detected, err := detectFormat(input)
	if err != nil {
		return imagecompression.Result{}, err
	}

	switch detected.format {
	case imagecompression.FormatJPEG:
		return c.compressJPEG(input)
	case imagecompression.FormatPNG:
		return c.compressPNG(input)
	case imagecompression.FormatWebP:
		return c.compressWebP(input)
	case imagecompression.FormatAVIF:
		return c.compressAVIF(input)
	case imagecompression.FormatHEIF:
		return c.compressHEIF(input, detected.contentType)
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

func (c Compressor) compressJPEG(input []byte) (imagecompression.Result, error) {
	cfg, err := jpeg.DecodeConfig(bytes.NewReader(input))
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	if err := validateDimensions(cfg, c.maxDecodedPixels()); err != nil {
		return imagecompression.Result{}, err
	}

	img, err := jpeg.Decode(bytes.NewReader(input))
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	img = applyOrientation(img, readEXIFOrientation(input))

	var output bytes.Buffer
	if err := jpeg.Encode(&output, img, &jpeg.Options{Quality: c.jpegQuality()}); err != nil {
		return imagecompression.Result{}, fmt.Errorf("encode jpeg: %w", err)
	}

	return staticResult(output.Bytes(), imagecompression.FormatJPEG, "image/jpeg", img), nil
}

func (c Compressor) compressPNG(input []byte) (imagecompression.Result, error) {
	cfg, err := png.DecodeConfig(bytes.NewReader(input))
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	if err := validateDimensions(cfg, c.maxDecodedPixels()); err != nil {
		return imagecompression.Result{}, err
	}

	img, err := png.Decode(bytes.NewReader(input))
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	var output bytes.Buffer
	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	if err := encoder.Encode(&output, img); err != nil {
		return imagecompression.Result{}, fmt.Errorf("encode png: %w", err)
	}

	return staticResult(output.Bytes(), imagecompression.FormatPNG, "image/png", img), nil
}

func (c Compressor) compressGIF(input []byte) (imagecompression.Result, error) {
	cfg, err := gif.DecodeConfig(bytes.NewReader(input))
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	if err := validateDimensions(cfg, c.maxDecodedPixels()); err != nil {
		return imagecompression.Result{}, err
	}

	img, err := gif.DecodeAll(bytes.NewReader(input))
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}
	if len(img.Image) == 0 {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	if len(img.Image) > 1 {
		if err := validateAnimatedDimensions(cfg.Width, cfg.Height, len(img.Image), c.maxAnimatedFramePixels()); err != nil {
			return imagecompression.Result{}, err
		}
	}

	var output bytes.Buffer
	if err := gif.EncodeAll(&output, img); err != nil {
		return imagecompression.Result{}, fmt.Errorf("encode gif: %w", err)
	}

	return imagecompression.Result{
		Data:        output.Bytes(),
		Format:      imagecompression.FormatGIF,
		ContentType: "image/gif",
		Width:       cfg.Width,
		Height:      cfg.Height,
		Animated:    len(img.Image) > 1,
		FrameCount:  len(img.Image),
	}, nil
}

func (c Compressor) compressBMP(input []byte) (imagecompression.Result, error) {
	cfg, err := bmp.DecodeConfig(bytes.NewReader(input))
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	if err := validateDimensions(cfg, c.maxDecodedPixels()); err != nil {
		return imagecompression.Result{}, err
	}

	img, err := bmp.Decode(bytes.NewReader(input))
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	var output bytes.Buffer
	if err := bmp.Encode(&output, img); err != nil {
		return imagecompression.Result{}, fmt.Errorf("encode bmp: %w", err)
	}

	return staticResult(output.Bytes(), imagecompression.FormatBMP, "image/bmp", img), nil
}

func (c Compressor) compressTIFF(input []byte) (imagecompression.Result, error) {
	cfg, err := tiff.DecodeConfig(bytes.NewReader(input))
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	if err := validateDimensions(cfg, c.maxDecodedPixels()); err != nil {
		return imagecompression.Result{}, err
	}

	img, err := tiff.Decode(bytes.NewReader(input))
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	var output bytes.Buffer
	if err := tiff.Encode(&output, img, &tiff.Options{Compression: tiff.Deflate, Predictor: true}); err != nil {
		return imagecompression.Result{}, fmt.Errorf("encode tiff: %w", err)
	}

	return staticResult(output.Bytes(), imagecompression.FormatTIFF, "image/tiff", img), nil
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
		return DefaultMaxDecodedPixels
	}

	return c.MaxDecodedPixels
}

func (c Compressor) maxAnimatedFramePixels() int {
	if c.MaxAnimatedFramePixels <= 0 {
		return DefaultMaxAnimatedFramePixels
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
