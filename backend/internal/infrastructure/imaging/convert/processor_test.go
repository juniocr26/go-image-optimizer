package convert

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageconversion"
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageprocessing"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/imaging"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/imaging/resize"
	"hash/crc32"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"testing"
)

func fixture(t *testing.T) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 13, 7))
	img.Set(1, 1, color.NRGBA{R: 255, A: 128})
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}
func TestTargets(t *testing.T) {
	input := fixture(t)
	uc := imageconversion.NewUseCase(Processor{})
	for _, target := range imageconversion.Targets {
		if target == "png" {
			continue
		}
		t.Run(string(target), func(t *testing.T) {
			result, err := uc.Execute(context.Background(), input, target)
			if err != nil {
				t.Fatal(err)
			}
			detected, err := imaging.DetectFormat(result.Data)
			if err != nil || detected.Format != target || detected.ContentType != result.ContentType {
				t.Fatalf("wrong output: %+v %v", detected, err)
			}
			decoded, err := (resize.Decoder{}).Decode(result.Data)
			if err != nil {
				t.Fatal(err)
			}
			info := decoded.Info()
			if info.Width != 13 || info.Height != 7 {
				t.Fatal(info)
			}
			pixels := decoded.(interface{ Pixels() image.Image }).Pixels()
			r, g, b, a := pixels.At(0, 0).RGBA()
			if target == "jpeg" || target == "bmp" || target == "heif" {
				if r < 60000 || g < 60000 || b < 60000 || a != 65535 {
					t.Fatalf("expected white: %d %d %d %d", r, g, b, a)
				}
			} else if a > 1000 {
				t.Fatalf("alpha lost: %d", a)
			}
			_, _, _, partial := pixels.At(1, 1).RGBA()
			if target == "webp" || target == "avif" || target == "tiff" {
				if partial < 32000 || partial > 34000 {
					t.Fatalf("partial alpha lost: %d", partial)
				}
			} else if partial != 65535 {
				t.Fatalf("opaque/binary destination alpha: %d", partial)
			}

		})
	}
	// A JPEG source exercises PNG output and confirms a larger output is retained.
	jpegResult, err := uc.Execute(context.Background(), input, "jpeg")
	if err != nil {
		t.Fatal(err)
	}
	pngResult, err := uc.Execute(context.Background(), jpegResult.Data, "png")
	if err != nil || len(pngResult.Data) == 0 || pngResult.Format != "png" {
		t.Fatalf("PNG: %v", err)
	}
	// PNG cannot be its own destination: use an alpha-capable TIFF source.
	tiffResult, err := uc.Execute(context.Background(), input, "tiff")
	if err != nil {
		t.Fatal(err)
	}
	alphaPNG, err := uc.Execute(context.Background(), tiffResult.Data, "png")
	if err != nil {
		t.Fatal(err)
	}
	pixels, err := png.Decode(bytes.NewReader(alphaPNG.Data))
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, alpha := pixels.At(1, 1).RGBA()
	if alpha < 32000 || alpha > 34000 {
		t.Fatalf("PNG partial alpha lost: %d", alpha)
	}

}
func TestRejections(t *testing.T) {
	uc := imageconversion.NewUseCase(Processor{})
	input := fixture(t)
	for _, target := range []imageprocessing.Format{"png", "jpg", "invalid", "PNG", ""} {
		if _, err := uc.Execute(context.Background(), input, target); !errors.Is(err, imageconversion.ErrInvalidTarget) {
			t.Fatalf("%s: %v", target, err)
		}
	}
	if _, err := uc.Execute(context.Background(), []byte("invalid"), "webp"); err == nil {
		t.Fatal("malformed accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := uc.Execute(ctx, input, "webp"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	frame := image.NewPaletted(image.Rect(0, 0, 2, 2), color.Palette{color.Black, color.White})
	var out bytes.Buffer
	if err := gif.EncodeAll(&out, &gif.GIF{Image: []*image.Paletted{frame, frame}, Delay: []int{2, 3}, LoopCount: 2}); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.Execute(context.Background(), out.Bytes(), "webp"); !errors.Is(err, imageprocessing.ErrUnsupportedVariant) {
		t.Fatal(err)
	}
}

func TestConversionJPEGOrientation(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 40, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 40; x++ {
			if x < 20 {
				img.Set(x, y, color.RGBA{R: 255, A: 255})
			} else {
				img.Set(x, y, color.RGBA{B: 255, A: 255})
			}
		}
	}
	var raw bytes.Buffer
	if err := jpeg.Encode(&raw, img, &jpeg.Options{Quality: 100}); err != nil {
		t.Fatal(err)
	}
	// Little-endian EXIF IFD with orientation 6 (90 degrees clockwise).
	payload := []byte{'E', 'x', 'i', 'f', 0, 0, 'I', 'I', 42, 0, 8, 0, 0, 0, 1, 0, 0x12, 1, 3, 0, 1, 0, 0, 0, 6, 0, 0, 0, 0, 0, 0, 0}
	input := append([]byte{}, raw.Bytes()[:2]...)
	input = append(input, 0xff, 0xe1, byte((len(payload)+2)>>8), byte(len(payload)+2))
	input = append(input, payload...)
	input = append(input, raw.Bytes()[2:]...)
	uc := imageconversion.NewUseCase(Processor{})
	info, err := uc.Inspect(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if info.Width != 20 || info.Height != 40 {
		t.Fatalf("oriented dimensions: %+v", info)
	}
	result, err := uc.Execute(context.Background(), input, "png")
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := png.Decode(bytes.NewReader(result.Data))
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Bounds().Dx() != 20 || decoded.Bounds().Dy() != 40 {
		t.Fatalf("wrong oriented output: %v", decoded.Bounds())
	}
	r, _, b, _ := decoded.At(5, 2).RGBA()
	if r < 50000 || b > 10000 {
		t.Fatal("orientation did not move left half to top")
	}
	r, _, b, _ = decoded.At(5, 35).RGBA()
	if b < 50000 || r > 10000 {
		t.Fatal("orientation did not move right half to bottom")
	}
	if imaging.ReadEXIFOrientation(result.Data) > 1 {
		t.Fatal("retained stale EXIF orientation")
	}
}

func TestVariantsAndPixelLimits(t *testing.T) {
	uc := imageconversion.NewUseCase(Processor{})
	data := fixture(t)
	chunk := make([]byte, 20)
	binary.BigEndian.PutUint32(chunk[:4], 8)
	copy(chunk[4:8], "acTL")
	binary.BigEndian.PutUint32(chunk[8:12], 2)
	binary.BigEndian.PutUint32(chunk[16:], crc32.ChecksumIEEE(chunk[4:16]))
	apng := append([]byte{}, data[:33]...)
	apng = append(apng, chunk...)
	apng = append(apng, data[33:]...)
	multipage := []byte{'I', 'I', 42, 0, 8, 0, 0, 0, 0, 0, 14, 0, 0, 0, 0, 0, 0, 0}
	for _, input := range [][]byte{apng, multipage} {
		if _, err := uc.Execute(context.Background(), input, "webp"); !errors.Is(err, imageprocessing.ErrUnsupportedVariant) {
			t.Fatal(err)
		}
	}
	oversized := append([]byte{}, data...)
	binary.BigEndian.PutUint32(oversized[16:20], 8001)
	binary.BigEndian.PutUint32(oversized[20:24], 4000)
	binary.BigEndian.PutUint32(oversized[29:33], crc32.ChecksumIEEE(oversized[12:29]))
	if _, err := uc.Execute(context.Background(), oversized, "webp"); !errors.Is(err, imageprocessing.ErrImageTooLarge) {
		t.Fatal(err)
	}
	frame := image.NewPaletted(image.Rect(0, 0, 1, 1), color.Palette{color.Black, color.White})
	var out bytes.Buffer
	if err := gif.EncodeAll(&out, &gif.GIF{Image: []*image.Paletted{frame, frame, frame}, Delay: []int{1, 1, 1}}); err != nil {
		t.Fatal(err)
	}
	bomb := out.Bytes()
	binary.LittleEndian.PutUint16(bomb[6:8], 8000)
	binary.LittleEndian.PutUint16(bomb[8:10], 4000)
	if _, err := uc.Execute(context.Background(), bomb, "webp"); !errors.Is(err, imageprocessing.ErrAnimationTooLarge) {
		t.Fatal(err)
	}
}
