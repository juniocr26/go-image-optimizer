package imaging

import (
	"image"
	"image/color"
	"testing"
)

func TestApplyOrientationCoversEXIFValues(t *testing.T) {
	source := coordinateImage(3, 2)

	tests := []struct {
		name        string
		orientation uint16
		width       int
		height      int
		mapPoint    func(x, y int) (int, int)
	}{
		{name: "1", orientation: 1, width: 3, height: 2, mapPoint: func(x, y int) (int, int) { return x, y }},
		{name: "2", orientation: 2, width: 3, height: 2, mapPoint: func(x, y int) (int, int) { return 2 - x, y }},
		{name: "3", orientation: 3, width: 3, height: 2, mapPoint: func(x, y int) (int, int) { return 2 - x, 1 - y }},
		{name: "4", orientation: 4, width: 3, height: 2, mapPoint: func(x, y int) (int, int) { return x, 1 - y }},
		{name: "5", orientation: 5, width: 2, height: 3, mapPoint: func(x, y int) (int, int) { return y, x }},
		{name: "6", orientation: 6, width: 2, height: 3, mapPoint: func(x, y int) (int, int) { return 1 - y, x }},
		{name: "7", orientation: 7, width: 2, height: 3, mapPoint: func(x, y int) (int, int) { return 1 - y, 2 - x }},
		{name: "8", orientation: 8, width: 2, height: 3, mapPoint: func(x, y int) (int, int) { return y, 2 - x }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oriented := ApplyOrientation(source, tt.orientation)
			assertOrientationDimensions(t, oriented, tt.width, tt.height)

			for y := 0; y < 2; y++ {
				for x := 0; x < 3; x++ {
					gotX, gotY := tt.mapPoint(x, y)
					expected := color.NRGBAModel.Convert(source.At(x, y)).(color.NRGBA)
					actual := color.NRGBAModel.Convert(oriented.At(gotX, gotY)).(color.NRGBA)
					if actual != expected {
						t.Fatalf("orientation %d mapped %d,%d to wrong color at %d,%d: expected %#v, got %#v", tt.orientation, x, y, gotX, gotY, expected, actual)
					}
				}
			}
		})
	}
}

func coordinateImage(width, height int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8(30 + x*50),
				G: uint8(40 + y*70),
				B: uint8(20 + x*10 + y*30),
				A: 255,
			})
		}
	}

	return img
}

func assertOrientationDimensions(t *testing.T, img image.Image, width, height int) {
	t.Helper()

	bounds := img.Bounds()
	if bounds.Dx() != width || bounds.Dy() != height {
		t.Fatalf("expected dimensions %dx%d, got %dx%d", width, height, bounds.Dx(), bounds.Dy())
	}
}
