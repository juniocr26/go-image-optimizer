package compress

import (
	"bytes"
	"fmt"
	"image/jpeg"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/imaging"
)

func (c Compressor) compressJPEG(input []byte) (imagecompression.Result, error) {
	cfg, err := jpeg.DecodeConfig(bytes.NewReader(input))
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	if err := imaging.ValidateDimensions(cfg, c.maxDecodedPixels()); err != nil {
		return imagecompression.Result{}, err
	}

	img, err := jpeg.Decode(bytes.NewReader(input))
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	img = imaging.ApplyOrientation(img, imaging.ReadEXIFOrientation(input))

	var output bytes.Buffer
	if err := jpeg.Encode(&output, img, &jpeg.Options{Quality: c.jpegQuality()}); err != nil {
		return imagecompression.Result{}, fmt.Errorf("encode jpeg: %w", err)
	}

	return staticResult(output.Bytes(), imagecompression.FormatJPEG, "image/jpeg", img), nil
}
