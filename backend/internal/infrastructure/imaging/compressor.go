package imaging

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"

	gometadata "github.com/FlavioCFOliveira/GoMetadata"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
)

const (
	// DefaultJPEGQuality keeps JPEG output visually close to the source while
	// still allowing meaningful size reduction for many camera/exported images.
	DefaultJPEGQuality = 82

	// DefaultMaxDecodedPixels limits decoded image memory growth. A 32 MP RGBA
	// image needs roughly 128 MiB before encoder overhead.
	DefaultMaxDecodedPixels = 32_000_000
)

type Compressor struct {
	JPEGQuality      int
	MaxDecodedPixels int
}

func NewCompressor() Compressor {
	return Compressor{
		JPEGQuality:      DefaultJPEGQuality,
		MaxDecodedPixels: DefaultMaxDecodedPixels,
	}
}

func (c Compressor) Compress(input []byte) (imagecompression.Result, error) {
	format, err := detectFormat(input)
	if err != nil {
		return imagecompression.Result{}, err
	}

	switch format {
	case imagecompression.FormatJPEG:
		return c.compressJPEG(input)
	case imagecompression.FormatPNG:
		return c.compressPNG(input)
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

	orientation := readEXIFOrientation(input)
	img = applyOrientation(img, orientation)

	var output bytes.Buffer

	if err := jpeg.Encode(
		&output,
		img,
		&jpeg.Options{Quality: c.jpegQuality()},
	); err != nil {
		return imagecompression.Result{}, fmt.Errorf("encode jpeg: %w", err)
	}

	bounds := img.Bounds()

	return imagecompression.Result{
		Data:        output.Bytes(),
		Format:      imagecompression.FormatJPEG,
		ContentType: "image/jpeg",
		Width:       bounds.Dx(),
		Height:      bounds.Dy(),
	}, nil
}

func applyOrientation(src image.Image, orientation uint16) image.Image {
	if orientation <= 1 || orientation > 8 {
		return src
	}

	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var dst *image.NRGBA

	if orientation >= 5 {
		dst = image.NewNRGBA(image.Rect(0, 0, height, width))
	} else {
		dst = image.NewNRGBA(image.Rect(0, 0, width, height))
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			color := src.At(bounds.Min.X+x, bounds.Min.Y+y)

			switch orientation {
			case 2:
				dst.Set(width-1-x, y, color)
			case 3:
				dst.Set(width-1-x, height-1-y, color)
			case 4:
				dst.Set(x, height-1-y, color)
			case 5:
				dst.Set(y, x, color)
			case 6:
				dst.Set(height-1-y, x, color)
			case 7:
				dst.Set(height-1-y, width-1-x, color)
			case 8:
				dst.Set(y, width-1-x, color)
			}
		}
	}

	return dst
}

func readEXIFOrientation(input []byte) uint16 {
	metadata, err := gometadata.Read(bytes.NewReader(input))
	if err != nil {
		return 1
	}

	orientation, ok := metadata.Orientation()
	if !ok {
		return 1
	}

	return orientation
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

	return imagecompression.Result{
		Data:        output.Bytes(),
		Format:      imagecompression.FormatPNG,
		ContentType: "image/png",
		Width:       cfg.Width,
		Height:      cfg.Height,
	}, nil
}

func detectFormat(input []byte) (imagecompression.Format, error) {
	switch http.DetectContentType(input) {
	case "image/jpeg":
		return imagecompression.FormatJPEG, nil
	case "image/png":
		return imagecompression.FormatPNG, nil
	default:
		return "", imagecompression.ErrUnsupportedFormat
	}
}

func validateDimensions(cfg image.Config, maxPixels int) error {
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return imagecompression.ErrInvalidImage
	}

	if cfg.Width > maxPixels/cfg.Height {
		return imagecompression.ErrImageTooLarge
	}

	return nil
}

func (c Compressor) jpegQuality() int {
	if c.JPEGQuality == 0 {
		return DefaultJPEGQuality
	}

	if c.JPEGQuality < 1 {
		return 1
	}

	if c.JPEGQuality > 100 {
		return 100
	}

	return c.JPEGQuality
}

func (c Compressor) maxDecodedPixels() int {
	if c.MaxDecodedPixels <= 0 {
		return DefaultMaxDecodedPixels
	}

	return c.MaxDecodedPixels
}
