package compress

import (
	"bytes"
	"fmt"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/imaging"
	"golang.org/x/image/tiff"
)

func (c Compressor) compressTIFF(input []byte) (imagecompression.Result, error) {
	cfg, err := tiff.DecodeConfig(bytes.NewReader(input))
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	if err := imaging.ValidateDimensions(cfg, c.maxDecodedPixels()); err != nil {
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
