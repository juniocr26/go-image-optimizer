package imaging

import (
	"bytes"
	"image"

	gometadata "github.com/FlavioCFOliveira/GoMetadata"
)

func ApplyOrientation(src image.Image, orientation uint16) image.Image {
	if orientation <= 1 || orientation > 8 {
		return src
	}

	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var dst *image.NRGBA
	if orientation >= 5 {
		dst = image.NewNRGBA(image.Rect(0, 0, height, width))
	} else {
		dst = image.NewNRGBA(image.Rect(0, 0, width, height))
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			color := src.At(bounds.Min.X+x, bounds.Min.Y+y)

			switch orientation {
			case 2:
				dst.Set(width-1-x, y, color)
			case 3:
				dst.Set(width-1-x, height-1-y, color)
			case 4:
				dst.Set(x, height-1-y, color)
			case 5:
				dst.Set(y, x, color)
			case 6:
				dst.Set(height-1-y, x, color)
			case 7:
				dst.Set(height-1-y, width-1-x, color)
			case 8:
				dst.Set(y, width-1-x, color)
			}
		}
	}

	return dst
}

func ReadEXIFOrientation(input []byte) uint16 {
	metadata, err := gometadata.Read(bytes.NewReader(input))
	if err != nil {
		return 1
	}

	orientation, ok := metadata.Orientation()
	if !ok {
		return 1
	}

	return orientation
}
