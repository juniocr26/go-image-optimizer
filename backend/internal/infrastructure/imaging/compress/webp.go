package compress

import (
	"bytes"
	"fmt"

	"github.com/deepteams/webp"
	"github.com/deepteams/webp/animation"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/imaging"
)

func (c Compressor) compressWebP(input []byte) (imagecompression.Result, error) {
	features, err := webp.GetFeatures(bytes.NewReader(input))
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	frameCount := max(1, features.FrameCount)
	if err := imaging.ValidateImageSize(features.Width, features.Height, c.maxDecodedPixels()); err != nil {
		return imagecompression.Result{}, err
	}

	if features.HasAnimation || frameCount > 1 {
		return c.compressAnimatedWebP(input, features.Width, features.Height, frameCount, features.LoopCount, features.Format == "lossless")
	}

	img, err := webp.Decode(bytes.NewReader(input))
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	var output bytes.Buffer
	if err := webp.Encode(&output, img, &webp.EncoderOptions{
		Lossless:     features.Format == "lossless",
		Quality:      float32(c.webpQuality()),
		Method:       4,
		Exact:        true,
		AlphaQuality: 100,
	}); err != nil {
		return imagecompression.Result{}, fmt.Errorf("encode webp: %w", err)
	}

	return staticResult(output.Bytes(), imagecompression.FormatWebP, "image/webp", img), nil
}

func (c Compressor) compressAnimatedWebP(input []byte, width, height, frameCount, loopCount int, lossless bool) (imagecompression.Result, error) {
	if err := imaging.ValidateAnimatedDimensions(width, height, frameCount, c.maxAnimatedFramePixels()); err != nil {
		return imagecompression.Result{}, err
	}

	anim, err := animation.Decode(bytes.NewReader(input))
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}
	if len(anim.Frames) == 0 {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}
	if err := anim.DecodeFrames(); err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	decoder, err := animation.NewAnimDecoder(anim)
	if err != nil {
		return imagecompression.Result{}, imagecompression.ErrInvalidImage
	}

	var output bytes.Buffer
	encoder := animation.NewEncoder(&output, width, height, &animation.EncodeOptions{
		LoopCount:       loopCount,
		BackgroundColor: anim.BackgroundColor,
		Quality:         c.webpQuality(),
		Lossless:        lossless,
	})
	if len(anim.ICC) > 0 {
		encoder.SetICCProfile(anim.ICC)
	}
	if len(anim.EXIF) > 0 {
		encoder.SetEXIF(anim.EXIF)
	}
	if len(anim.XMP) > 0 {
		encoder.SetXMP(anim.XMP)
	}

	for decoder.HasNext() {
		frame, duration, err := decoder.NextFrame()
		if err != nil {
			return imagecompression.Result{}, imagecompression.ErrInvalidImage
		}

		if err := encoder.AddFrame(frame, duration); err != nil {
			return imagecompression.Result{}, fmt.Errorf("encode webp frame: %w", err)
		}
	}

	if err := encoder.Close(); err != nil {
		return imagecompression.Result{}, fmt.Errorf("encode webp animation: %w", err)
	}

	return imagecompression.Result{
		Data:        output.Bytes(),
		Format:      imagecompression.FormatWebP,
		ContentType: "image/webp",
		Width:       width,
		Height:      height,
		Animated:    true,
		FrameCount:  frameCount,
	}, nil
}
