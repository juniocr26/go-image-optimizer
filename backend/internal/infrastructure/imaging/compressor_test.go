package imaging

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
)

func TestCompressorCompressesJPEGAndPreservesFormatAndDimensions(t *testing.T) {
	source := gradientImage(96, 64)
	input := encodeJPEG(t, source, 100)

	result, err := NewCompressor().Compress(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Format != imagecompression.FormatJPEG {
		t.Fatalf("expected JPEG format, got %q", result.Format)
	}

	if result.ContentType != "image/jpeg" {
		t.Fatalf("expected image/jpeg content type, got %q", result.ContentType)
	}

	decoded, format, err := image.Decode(bytes.NewReader(result.Data))
	if err != nil {
		t.Fatalf("compressed JPEG could not be decoded: %v", err)
	}

	if format != "jpeg" {
		t.Fatalf("expected decoded format jpeg, got %q", format)
	}

	assertDimensions(t, decoded, 96, 64)
}

func TestCompressorReducesDeterministicHighQualityJPEGFixture(t *testing.T) {
	source := detailedImage(120, 90)
	input := encodeJPEG(t, source, 100)

	result, err := NewCompressor().Compress(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result.Data) >= len(input) {
		t.Fatalf("expected high-quality JPEG fixture to shrink, original=%d optimized=%d", len(input), len(result.Data))
	}
}

func TestCompressorCompressesPNGLosslesslyAndPreservesFormatAndDimensions(t *testing.T) {
	source := gradientImage(80, 52)
	input := encodePNG(t, source, png.NoCompression)

	result, err := NewCompressor().Compress(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Format != imagecompression.FormatPNG {
		t.Fatalf("expected PNG format, got %q", result.Format)
	}

	if result.ContentType != "image/png" {
		t.Fatalf("expected image/png content type, got %q", result.ContentType)
	}

	decoded, format, err := image.Decode(bytes.NewReader(result.Data))
	if err != nil {
		t.Fatalf("compressed PNG could not be decoded: %v", err)
	}

	if format != "png" {
		t.Fatalf("expected decoded format png, got %q", format)
	}

	assertDimensions(t, decoded, 80, 52)
	assertPixelsEqual(t, source, decoded)
}

func TestCompressorRejectsUnsupportedInput(t *testing.T) {
	_, err := NewCompressor().Compress([]byte("plain text"))

	if !errors.Is(err, imagecompression.ErrUnsupportedFormat) {
		t.Fatalf("expected ErrUnsupportedFormat, got %v", err)
	}
}

func TestCompressorRejectsCorruptedImageContent(t *testing.T) {
	corruptedPNG := []byte("\x89PNG\r\n\x1a\nnot a valid png body")

	_, err := NewCompressor().Compress(corruptedPNG)

	if !errors.Is(err, imagecompression.ErrInvalidImage) {
		t.Fatalf("expected ErrInvalidImage, got %v", err)
	}
}

func TestCompressorRejectsImagesAboveDecodedPixelLimit(t *testing.T) {
	source := gradientImage(2, 2)
	input := encodePNG(t, source, png.DefaultCompression)
	compressor := Compressor{
		JPEGQuality:      DefaultJPEGQuality,
		MaxDecodedPixels: 3,
	}

	_, err := compressor.Compress(input)

	if !errors.Is(err, imagecompression.ErrImageTooLarge) {
		t.Fatalf("expected ErrImageTooLarge, got %v", err)
	}
}

func gradientImage(width, height int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8((x * 255) / max(1, width-1)),
				G: uint8((y * 255) / max(1, height-1)),
				B: uint8((x*y + x + y) % 256),
				A: 255,
			})
		}
	}

	return img
}

func detailedImage(width, height int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			v := (x*37 + y*53 + x*y*11) % 256
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8(v),
				G: uint8((v + x*7) % 256),
				B: uint8((v + y*13) % 256),
				A: 255,
			})
		}
	}

	return img
}

func encodeJPEG(t *testing.T, img image.Image, quality int) []byte {
	t.Helper()

	var buffer bytes.Buffer
	if err := jpeg.Encode(&buffer, img, &jpeg.Options{Quality: quality}); err != nil {
		t.Fatalf("failed to encode jpeg fixture: %v", err)
	}

	return buffer.Bytes()
}

func encodePNG(t *testing.T, img image.Image, level png.CompressionLevel) []byte {
	t.Helper()

	var buffer bytes.Buffer
	encoder := png.Encoder{CompressionLevel: level}
	if err := encoder.Encode(&buffer, img); err != nil {
		t.Fatalf("failed to encode png fixture: %v", err)
	}

	return buffer.Bytes()
}

func assertDimensions(t *testing.T, img image.Image, width, height int) {
	t.Helper()

	bounds := img.Bounds()
	if bounds.Dx() != width || bounds.Dy() != height {
		t.Fatalf("expected dimensions %dx%d, got %dx%d", width, height, bounds.Dx(), bounds.Dy())
	}
}

func assertPixelsEqual(t *testing.T, expected image.Image, actual image.Image) {
	t.Helper()

	expectedBounds := expected.Bounds()
	if expectedBounds.Dx() != actual.Bounds().Dx() || expectedBounds.Dy() != actual.Bounds().Dy() {
		t.Fatalf("image bounds differ: expected %v, got %v", expectedBounds, actual.Bounds())
	}

	for y := expectedBounds.Min.Y; y < expectedBounds.Max.Y; y++ {
		for x := expectedBounds.Min.X; x < expectedBounds.Max.X; x++ {
			expectedColor := color.NRGBAModel.Convert(expected.At(x, y)).(color.NRGBA)
			actualColor := color.NRGBAModel.Convert(actual.At(x, y)).(color.NRGBA)

			if expectedColor != actualColor {
				t.Fatalf("pixel mismatch at %d,%d: expected %#v, got %#v", x, y, expectedColor, actualColor)
			}
		}
	}
}
