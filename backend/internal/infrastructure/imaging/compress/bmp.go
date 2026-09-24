package compress

import (
	"bytes"
	"fmt"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/imaging"
	"golang.org/x/image/bmp"
)

func (c Compressor) compressBMP(input []byte) (imagecompression.Result, error) {
	cfg, err := bmp.DecodeConfig(bytes.NewReader(input))
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	if err := imaging.ValidateDimensions(cfg, c.maxDecodedPixels()); err != nil {
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
