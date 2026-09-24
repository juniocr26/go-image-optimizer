package compress

import (
	"bytes"
	"fmt"
	"image/gif"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/imaging"
)

func (c Compressor) compressGIF(input []byte) (imagecompression.Result, error) {
	cfg, err := gif.DecodeConfig(bytes.NewReader(input))
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	if err := imaging.ValidateDimensions(cfg, c.maxDecodedPixels()); err != nil {
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
		if err := imaging.ValidateAnimatedDimensions(cfg.Width, cfg.Height, len(img.Image), c.maxAnimatedFramePixels()); err != nil {
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
