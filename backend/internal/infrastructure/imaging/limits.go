package imaging

import (
	"image"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
)

const (
	// DefaultMaxDecodedPixels limits decoded image memory growth. A 32 MP RGBA
	// image needs roughly 128 MiB before encoder overhead.
	DefaultMaxDecodedPixels = 32_000_000

	// DefaultMaxAnimatedFramePixels caps canvas-sized animation work. The limit
	// is width * height * frame count, not bytes, so codecs can share it.
	DefaultMaxAnimatedFramePixels = 64_000_000
)

func validateDimensions(cfg image.Config, maxPixels int) error {
	return validateImageSize(cfg.Width, cfg.Height, maxPixels)
}

func validateImageSize(width, height, maxPixels int) error {
	if width <= 0 || height <= 0 {
		return imagecompression.ErrInvalidImage
	}

	if width > maxPixels/height {
		return imagecompression.ErrImageTooLarge
	}

	return nil
}

func validateAnimatedDimensions(width, height, frameCount, maxPixels int) error {
	if err := validateImageSize(width, height, DefaultMaxDecodedPixels); err != nil {
		return err
	}

	if frameCount <= 0 {
		return imagecompression.ErrInvalidImage
	}

	if width > maxPixels/height/frameCount {
		return imagecompression.ErrAnimationTooLarge
	}

	return nil
}
