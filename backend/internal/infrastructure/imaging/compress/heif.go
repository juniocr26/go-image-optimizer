package compress

import (
	"errors"
	"fmt"
	"image"
	"os"
	"strings"

	"github.com/strukturag/libheif/go/heif"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/imaging"
)

func (c Compressor) compressHEIF(input []byte, detectedContentType string) (imagecompression.Result, error) {
	ctx, err := heif.NewContext()
	if err != nil {
		return imagecompression.Result{}, fmt.Errorf("%w: %v", imagecompression.ErrCodecUnavailable, err)
	}

	if err := ctx.ReadFromMemory(input); err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	if ctx.GetNumberOfTopLevelImages() != 1 {
		return imagecompression.Result{}, imagecompression.ErrUnsupportedVariant
	}

	handle, err := ctx.GetPrimaryImageHandle()
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	if err := imaging.ValidateImageSize(handle.GetWidth(), handle.GetHeight(), c.maxDecodedPixels()); err != nil {
		return imagecompression.Result{}, err
	}

	options, err := heif.NewDecodingOptions()
	if err != nil {
		return imagecompression.Result{}, fmt.Errorf("%w: %v", imagecompression.ErrCodecUnavailable, err)
	}

	decoded, err := handle.DecodeImage(heif.ColorspaceRGB, heif.ChromaInterleavedRGBA, options)
	if err != nil {
		return imagecompression.Result{}, mapHEIFDecodeError(err)
	}

	img, err := decoded.GetImage()
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	rgba := toRGBA(img)
	outCtx, err := heif.EncodeFromImage(
		rgba,
		heif.CompressionHEVC,
		c.heifQuality(),
		heif.LosslessModeDisabled,
		heif.LoggingLevelNone,
	)
	if err != nil {
		return imagecompression.Result{}, mapHEIFEncodeError(err)
	}

	output, err := writeHEIFContextToBytes(outCtx)
	if err != nil {
		return imagecompression.Result{}, err
	}

	contentType := detectedContentType
	if detected, err := imaging.DetectFormat(output); err == nil && detected.Format == imagecompression.FormatHEIF {
		contentType = detected.ContentType
	}

	result := staticResult(output, imagecompression.FormatHEIF, contentType, rgba)
	return result, nil
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
		return fmt.Errorf("%w: %v", imagecompression.ErrCodecUnavailable, err)
	}

	return imagecompression.ErrInvalidImage
}

func mapHEIFEncodeError(err error) error {
	return fmt.Errorf("%w: %v", imagecompression.ErrCodecUnavailable, err)
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
