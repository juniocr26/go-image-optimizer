package imaging

import (
	"errors"
	"fmt"
	"image"
	"os"
	"strings"

	"github.com/strukturag/libheif/go/heif"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageprocessing"
)

func DecodeHEIF(input []byte, maxPixels int) (image.Image, error) {
	ctx, err := heif.NewContext()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", imageprocessing.ErrCodecUnavailable, err)
	}

	if err := ctx.ReadFromMemory(input); err != nil {
		return nil, imageprocessing.ErrInvalidImage
	}

	if ctx.GetNumberOfTopLevelImages() != 1 {
		return nil, imageprocessing.ErrUnsupportedVariant
	}

	handle, err := ctx.GetPrimaryImageHandle()
	if err != nil {
		return nil, imageprocessing.ErrInvalidImage
	}

	if err := ValidateImageSize(handle.GetWidth(), handle.GetHeight(), maxPixels); err != nil {
		return nil, err
	}

	options, err := heif.NewDecodingOptions()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", imageprocessing.ErrCodecUnavailable, err)
	}

	decoded, err := handle.DecodeImage(heif.ColorspaceRGB, heif.ChromaInterleavedRGBA, options)
	if err != nil {
		return nil, mapHEIFDecodeError(err)
	}

	img, err := decoded.GetImage()
	if err != nil {
		return nil, imageprocessing.ErrInvalidImage
	}

	return img, nil
}

func EncodeHEIF(img image.Image, quality int) ([]byte, error) {
	rgba := toRGBA(img)
	outCtx, err := heif.EncodeFromImage(
		rgba,
		heif.CompressionHEVC,
		quality,
		heif.LosslessModeDisabled,
		heif.LoggingLevelNone,
	)
	if err != nil {
		return nil, mapHEIFEncodeError(err)
	}

	output, err := writeHEIFContextToBytes(outCtx)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func toRGBA(img image.Image) *image.RGBA {
	bounds := img.Bounds()
	rgba := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x-bounds.Min.X, y-bounds.Min.Y, img.At(x, y))
		}
	}

	return rgba
}

func writeHEIFContextToBytes(ctx *heif.Context) ([]byte, error) {
	tmp, err := os.CreateTemp("", "go-image-optimizer-*.heic")
	if err != nil {
		return nil, fmt.Errorf("create heif temp file: %w", err)
	}

	name := tmp.Name()
	if err := tmp.Close(); err != nil {
		_ = os.Remove(name)
		return nil, fmt.Errorf("close heif temp file: %w", err)
	}
	defer os.Remove(name)

	if err := ctx.WriteToFile(name); err != nil {
		return nil, mapHEIFEncodeError(err)
	}

	output, err := os.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("read heif temp file: %w", err)
	}

	return output, nil
}

func mapHEIFDecodeError(err error) error {
	if isHEIFCodecError(err) {
		return fmt.Errorf("%w: %v", imageprocessing.ErrCodecUnavailable, err)
	}

	return imageprocessing.ErrInvalidImage
}

func mapHEIFEncodeError(err error) error {
	return fmt.Errorf("%w: %v", imageprocessing.ErrCodecUnavailable, err)
}

func isHEIFCodecError(err error) bool {
	var heifErr *heif.HeifError
	if errors.As(err, &heifErr) {
		message := strings.ToLower(heifErr.Message)
		return strings.Contains(message, "plugin") ||
			strings.Contains(message, "encoder") ||
			strings.Contains(message, "decoder") ||
			strings.Contains(message, "codec")
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "plugin") ||
		strings.Contains(message, "encoder") ||
		strings.Contains(message, "decoder") ||
		strings.Contains(message, "codec")
}
