package compress

import (
	"bytes"
	"fmt"
	"image/png"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/imaging"
)

func (c Compressor) compressPNG(input []byte) (imagecompression.Result, error) {
	cfg, err := png.DecodeConfig(bytes.NewReader(input))
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	if err := imaging.ValidateDimensions(cfg, c.maxDecodedPixels()); err != nil {
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
