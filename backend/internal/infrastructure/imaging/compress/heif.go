package compress

import (
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/imaging"
)

func (c Compressor) compressHEIF(input []byte, detectedContentType string) (imagecompression.Result, error) {
	img, err := imaging.DecodeHEIF(input, c.maxDecodedPixels())
	if err != nil {
		return imagecompression.Result{}, err
	}
	output, err := imaging.EncodeHEIF(img, c.heifQuality())
	if err != nil {
		return imagecompression.Result{}, err
	}
	contentType := detectedContentType
	if detected, err := imaging.DetectFormat(output); err == nil {
		contentType = detected.ContentType
	}
	return staticResult(output, imagecompression.FormatHEIF, contentType, img), nil
}
