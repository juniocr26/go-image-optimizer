package imaging

import (
	"bytes"
	"fmt"

	"github.com/gen2brain/avif"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
)

func (c Compressor) compressAVIF(input []byte) (imagecompression.Result, error) {
	cfg, err := avif.DecodeConfig(bytes.NewReader(input))
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	if err := validateDimensions(cfg, c.maxDecodedPixels()); err != nil {
		return imagecompression.Result{}, err
	}

	anim, err := avif.DecodeAll(bytes.NewReader(input), avif.Options{AutoRotate: true})
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}
	if len(anim.Image) == 0 {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	frameCount := len(anim.Image)
	firstBounds := anim.Image[0].Bounds()
	width := firstBounds.Dx()
	height := firstBounds.Dy()

	if frameCount > 1 {
		if err := validateAnimatedDimensions(width, height, frameCount, c.maxAnimatedFramePixels()); err != nil {
			return imagecompression.Result{}, err
		}
	}

	options := avif.Options{
		Quality:      c.avifQuality(),
		QualityAlpha: 100,
		Speed:        6,
	}

	var output bytes.Buffer
	if frameCount > 1 {
		if err := avif.EncodeAll(&output, anim, options); err != nil {
			return imagecompression.Result{}, fmt.Errorf("encode avif animation: %w", err)
		}
	} else if err := avif.Encode(&output, anim.Image[0], options); err != nil {
		return imagecompression.Result{}, fmt.Errorf("encode avif: %w", err)
	}

	return imagecompression.Result{
		Data:        output.Bytes(),
		Format:      imagecompression.FormatAVIF,
		ContentType: "image/avif",
		Width:       width,
		Height:      height,
		Animated:    frameCount > 1,
		FrameCount:  frameCount,
	}, nil
}
