package convert

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/gif"
	"os"
	"path/filepath"
	"testing"

	"github.com/gen2brain/avif"
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageconversion"
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageprocessing"
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageresize"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/imaging"
)

// Decode with codec APIs rather than the processing Decoder under test.
func decodeRealOutput(t *testing.T, data []byte, format imageprocessing.Format) image.Image {
	t.Helper()
	var img image.Image
	var err error
	switch format {
	case "heif":
		img, err = imaging.DecodeHEIF(data, imageresize.MaxPixels)
	case "avif":
		img, err = avif.Decode(bytes.NewReader(data), avif.Options{AutoRotate: true})
	default:
		img, _, err = image.Decode(bytes.NewReader(data))
	}
	if err != nil {
		t.Fatal(err)
	}
	return img
}

func TestConversionRealFixtures(t *testing.T) {
	dir := os.Getenv("REAL_IMAGE_FIXTURES_DIR")
	if dir == "" {
		dir = "../../../../../storage/testdata/images"
	}
	// Explicit inventory makes additions to the fixture directory require a policy.
	sources := []struct {
		file   string
		format imageprocessing.Format
	}{
		{"sample.jpg", "jpeg"}, {"sample.png", "png"}, {"sample.webp", "webp"},
		{"sample.avif", "avif"}, {"sample.heic", "heif"}, {"sample.heif", "heif"},
		{"sample.gif", "gif"}, {"sample.bmp", "bmp"}, {"sample.tiff", "tiff"},
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != len(sources) {
		t.Fatalf("fixture inventory changed: %d files, %d policies", len(files), len(sources))
	}
	uc := imageconversion.NewUseCase(Processor{})
	for _, src := range sources {
		t.Run(src.file, func(t *testing.T) {
			input, err := os.ReadFile(filepath.Join(dir, src.file))
			if err != nil {
				t.Fatal(err)
			}
			original := append([]byte(nil), input...)
			img := decodeRealOutput(t, input, src.format)
			if src.format != "avif" && src.format != "heif" {
				img = imaging.ApplyOrientation(img, imaging.ReadEXIFOrientation(input))
			}
			width, height := img.Bounds().Dx(), img.Bounds().Dy()
			info, err := uc.Inspect(context.Background(), input)
			if err != nil {
				t.Fatal(err)
			}
			if info.Format != src.format || info.Width != width || info.Height != height {
				t.Fatalf("inspection disagrees with codec: %+v versus %dx%d", info, width, height)
			}
			// All current physical fixtures are static. Animation rejection has dedicated synthetic tests.
			if info.FrameCount != 1 {
				t.Fatalf("fixture became animated; add explicit rejection policy: %+v", info)
			}
			t.Logf("source %s %dx%d %d bytes", src.format, width, height, len(input))
			for _, target := range imageconversion.Targets {
				name := string(target)
				if target == src.format {
					name += " rejected same family"
				}
				t.Run(name, func(t *testing.T) {
					result, err := uc.Execute(context.Background(), input, target)
					if target == src.format {
						if !errors.Is(err, imageconversion.ErrInvalidTarget) {
							t.Fatalf("same-family target: %v", err)
						}
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					if len(result.Data) == 0 || result.SourceFormat != src.format || result.Format != target || result.Width != width || result.Height != height || result.FrameCount != 1 || result.Animated {
						t.Fatalf("invalid result: %+v", result)
					}
					detected, err := imaging.DetectFormat(result.Data)
					if err != nil || detected.Format != target || detected.ContentType != result.ContentType {
						t.Fatalf("detected format/MIME: %+v %v", detected, err)
					}
					decoded := decodeRealOutput(t, result.Data, target)
					if decoded.Bounds().Dx() != width || decoded.Bounds().Dy() != height {
						t.Fatalf("decoded bounds %v want %dx%d", decoded.Bounds(), width, height)
					}
					if target == "gif" {
						animation, err := gif.DecodeAll(bytes.NewReader(result.Data))
						if err != nil {
							t.Fatal(err)
						}
						if len(animation.Image) != 1 {
							t.Fatal("static conversion produced animation")
						}
					}
					// Probe a fully transparent source pixel where present. Lossy colors are
					// not compared exactly; opaque targets must use documented white compositing.
					checked := false
					for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y && !checked; y++ {
						for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
							_, _, _, alpha := img.At(x, y).RGBA()
							if alpha != 0 {
								continue
							}
							r, g, b, a := decoded.At(x-img.Bounds().Min.X, y-img.Bounds().Min.Y).RGBA()
							if target == "jpeg" || target == "bmp" || target == "heif" {
								if a != 65535 || r < 55000 || g < 55000 || b < 55000 {
									t.Fatalf("white compositing: %d %d %d %d", r, g, b, a)
								}
							} else if a > 1000 {
								t.Fatalf("transparent pixel lost: %d", a)
							}
							checked = true
							break
						}
					}
					if !bytes.Equal(input, original) {
						t.Fatal("source bytes modified")
					}
				})
			}
		})
	}
}
