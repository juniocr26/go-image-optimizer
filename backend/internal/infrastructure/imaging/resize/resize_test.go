package resize

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/deepteams/webp"
	"github.com/deepteams/webp/animation"
	"github.com/gen2brain/avif"
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageprocessing"
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageresize"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/imaging"
)

func TestResizeRealFixtures(t *testing.T) {
	dir := os.Getenv("REAL_IMAGE_FIXTURES_DIR")
	if dir == "" {
		dir = "../../../../../storage/testdata/images"
	}
	uc := imageresize.NewUseCase(Decoder{})
	for _, ext := range []string{"jpg", "png", "webp", "avif", "heic", "heif", "gif", "bmp", "tiff"} {
		t.Run(ext, func(t *testing.T) {
			input, err := os.ReadFile(filepath.Join(dir, "sample."+ext))
			if err != nil {
				t.Fatal(err)
			}
			info, err := uc.Inspect(context.Background(), input)
			if err != nil {
				t.Fatal(err)
			}
			unchanged, err := uc.Execute(context.Background(), input, imageresize.Options{Mode: "pixels", Width: info.Width, Height: info.Height, Axis: "width"})
			if err != nil || !bytes.Equal(unchanged.Data, input) {
				t.Fatalf("same-size input must retain encoded bytes: %v", err)
			}
			out, err := uc.Execute(context.Background(), input, imageresize.Options{Mode: "percentage", Reduction: 50})
			if err != nil {
				t.Fatal(err)
			}
			w, h := max(1, (info.Width+1)/2), max(1, (info.Height+1)/2)
			if out.Width != w || out.Height != h || out.OriginalWidth != info.Width || out.OriginalHeight != info.Height || out.Format != info.Format || len(out.Data) == 0 {
				t.Fatalf("bad resize result: %dx%d", out.Width, out.Height)
			}
			detected, err := imaging.DetectFormat(out.Data)
			if err != nil || detected.Format != info.Format || detected.ContentType != out.ContentType {
				t.Fatalf("format/MIME: %+v %v", detected, err)
			}
			var decoded image.Image
			if ext == "heic" || ext == "heif" {
				decoded, err = imaging.DecodeHEIF(out.Data, imageresize.MaxPixels)
			} else if ext == "avif" {
				decoded, err = avif.Decode(bytes.NewReader(out.Data), avif.Options{AutoRotate: true})
			} else {
				decoded, _, err = image.Decode(bytes.NewReader(out.Data))
			}
			if err != nil {
				t.Fatal(err)
			}
			if decoded.Bounds().Dx() != w || decoded.Bounds().Dy() != h {
				t.Fatalf("actual output: %v want %dx%d", decoded.Bounds(), w, h)
			}
		})
	}
}

func TestResizePreservesTransparency(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 10; x < 20; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 200, A: 128})
		}
	}
	for _, format := range []string{"png", "webp"} {
		t.Run(format, func(t *testing.T) {
			var input bytes.Buffer
			if format == "png" {
				if err := png.Encode(&input, img); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := webp.Encode(&input, img, &webp.EncoderOptions{Lossless: true, Exact: true}); err != nil {
					t.Fatal(err)
				}
			}
			out, err := imageresize.NewUseCase(Decoder{}).Execute(context.Background(), input.Bytes(), imageresize.Options{Mode: "percentage", Reduction: 50})
			if err != nil {
				t.Fatal(err)
			}
			decoded, _, err := image.Decode(bytes.NewReader(out.Data))
			if err != nil {
				t.Fatal(err)
			}
			_, _, _, transparent := decoded.At(0, 0).RGBA()
			_, _, _, partial := decoded.At(9, 9).RGBA()
			if transparent != 0 || partial < 32000 || partial > 34000 {
				t.Fatalf("alpha lost: %d %d", transparent, partial)
			}
		})
	}
}

func TestResizeJPEGOrientation(t *testing.T) {
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
	uc := imageresize.NewUseCase(Decoder{})
	info, err := uc.Inspect(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if info.Width != 20 || info.Height != 40 {
		t.Fatalf("oriented dimensions: %+v", info)
	}
	result, err := uc.Execute(context.Background(), input, imageresize.Options{Mode: "percentage", Reduction: 50})
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := jpeg.Decode(bytes.NewReader(result.Data))
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Bounds().Dx() != 10 || decoded.Bounds().Dy() != 20 {
		t.Fatalf("wrong oriented output: %v", decoded.Bounds())
	}
	r, _, b, _ := decoded.At(5, 2).RGBA()
	if r < 50000 || b > 10000 {
		t.Fatal("orientation did not move left half to top")
	}
	r, _, b, _ = decoded.At(5, 17).RGBA()
	if b < 50000 || r > 10000 {
		t.Fatal("orientation did not move right half to bottom")
	}
	if imaging.ReadEXIFOrientation(result.Data) > 1 {
		t.Fatal("retained stale EXIF orientation")
	}
}

func TestResizeGIFPartialFramesAndDisposal(t *testing.T) {
	pal := color.Palette{color.Transparent, color.RGBA{R: 255, A: 255}, color.RGBA{B: 255, A: 255}, color.RGBA{G: 255, A: 255}}
	first := image.NewPaletted(image.Rect(0, 0, 20, 20), pal)
	for y := 0; y < 20; y++ {
		for x := 0; x < 10; x++ {
			first.SetColorIndex(x, y, 1)
		}
	}
	second := image.NewPaletted(image.Rect(10, 0, 20, 20), pal)
	for i := range second.Pix {
		second.Pix[i] = 2
	}
	third := image.NewPaletted(image.Rect(0, 0, 4, 4), pal)
	for i := range third.Pix {
		third.Pix[i] = 3
	}
	for _, disposal := range []byte{gif.DisposalPrevious, gif.DisposalBackground} {
		t.Run(string(rune('0'+disposal)), func(t *testing.T) {
			anim := &gif.GIF{Image: []*image.Paletted{first, second, third}, Delay: []int{7, 11, 13}, Disposal: []byte{gif.DisposalNone, disposal, gif.DisposalNone}, LoopCount: 3, Config: image.Config{Width: 20, Height: 20, ColorModel: pal}}
			var input bytes.Buffer
			if err := gif.EncodeAll(&input, anim); err != nil {
				t.Fatal(err)
			}
			out, err := imageresize.NewUseCase(Decoder{}).Execute(context.Background(), input.Bytes(), imageresize.Options{Mode: "percentage", Reduction: 50})
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := gif.DecodeAll(bytes.NewReader(out.Data))
			if err != nil {
				t.Fatal(err)
			}
			if len(decoded.Image) != 3 || decoded.LoopCount != 3 || decoded.Config.Width != 10 || decoded.Config.Height != 10 {
				t.Fatal("lost animation metadata")
			}
			for i, delay := range []int{7, 11, 13} {
				if decoded.Delay[i] != delay {
					t.Fatal("lost delay")
				}
			}
			r, _, _, a := decoded.Image[1].At(1, 8).RGBA()
			if r < 50000 || a == 0 {
				t.Fatal("partial frame lost previous canvas")
			}
			_, _, b, a := decoded.Image[1].At(8, 8).RGBA()
			if b < 50000 || a == 0 {
				t.Fatal("partial frame was not drawn")
			}
			_, _, _, a = decoded.Image[2].At(8, 8).RGBA()
			if a != 0 {
				t.Fatal("disposal was not applied")
			}
		})
	}
}

func TestResizeAnimatedWebPAndAVIF(t *testing.T) {
	one := image.NewRGBA(image.Rect(0, 0, 20, 12))
	two := image.NewRGBA(one.Bounds())
	for y := 0; y < 12; y++ {
		for x := 0; x < 20; x++ {
			one.Set(x, y, color.RGBA{R: 255, A: 255})
			two.Set(x, y, color.RGBA{B: 255, A: 255})
		}
	}
	for _, format := range []string{"webp", "avif"} {
		t.Run(format, func(t *testing.T) {
			var input bytes.Buffer
			if format == "webp" {
				enc := animation.NewEncoder(&input, 20, 12, &animation.EncodeOptions{LoopCount: 3, Quality: 82, Lossless: true})
				if err := enc.AddFrame(one, 80*time.Millisecond); err != nil {
					t.Fatal(err)
				}
				if err := enc.AddFrame(two, 120*time.Millisecond); err != nil {
					t.Fatal(err)
				}
				if err := enc.Close(); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := avif.EncodeAll(&input, &avif.AVIF{Image: []image.Image{one, two}, Delay: []float64{0.08, 0.12}, LoopCount: 3}, avif.Options{Quality: 60, Speed: 6}); err != nil {
					t.Fatal(err)
				}
			}
			out, err := imageresize.NewUseCase(Decoder{}).Execute(context.Background(), input.Bytes(), imageresize.Options{Mode: "percentage", Reduction: 50})
			if format == "avif" {
				if !errors.Is(err, imageprocessing.ErrUnsupportedVariant) {
					t.Fatalf("animated AVIF must be rejected without changing its loop: %v", err)
				}
				return
			}

			if err != nil {
				t.Fatal(err)
			}
			if !out.Animated || out.FrameCount != 2 || out.Width != 10 || out.Height != 6 {
				t.Fatalf("animation output: %+v", out.Result)
			}
			if format == "webp" {
				anim, err := animation.Decode(bytes.NewReader(out.Data))
				if err != nil {
					t.Fatal(err)
				}
				if len(anim.Frames) != 2 || anim.Frames[0].Duration != 80*time.Millisecond || anim.Frames[1].Duration != 120*time.Millisecond || anim.LoopCount != 3 {
					t.Fatalf("WebP timing/loop changed: count=%d loop=%d durations=%d,%d", len(anim.Frames), anim.LoopCount, anim.Frames[0].Duration, anim.Frames[1].Duration)
				}
			}
		})
	}
}

func TestResizeRejectsBadInputAndGIFBomb(t *testing.T) {
	for _, input := range [][]byte{[]byte("not an image"), {0xff, 0xd8, 0xff}, {0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}, []byte("GIF89a")} {
		if _, err := (Decoder{}).Decode(input); err == nil {
			t.Fatal("accepted invalid input")
		}
	}
	pal := color.Palette{color.Black, color.White}
	frame := image.NewPaletted(image.Rect(0, 0, 1, 1), pal)
	var input bytes.Buffer
	if err := gif.EncodeAll(&input, &gif.GIF{Image: []*image.Paletted{frame, frame, frame}, Delay: []int{1, 1, 1}}); err != nil {
		t.Fatal(err)
	}
	data := input.Bytes()
	binary.LittleEndian.PutUint16(data[6:8], 8000)
	binary.LittleEndian.PutUint16(data[8:10], 4000)
	_, err := (Decoder{}).Decode(data)
	if !errors.Is(err, imageprocessing.ErrAnimationTooLarge) {
		t.Fatalf("expected pre-decode frame limit, got %v", err)
	}
}

func TestResizeRejectsFlatteningVariants(t *testing.T) {
	var pngData bytes.Buffer
	if err := png.Encode(&pngData, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	// APNG animation-control chunk immediately after IHDR.
	chunk := make([]byte, 20)
	binary.BigEndian.PutUint32(chunk[:4], 8)
	copy(chunk[4:8], "acTL")
	binary.BigEndian.PutUint32(chunk[8:12], 2)
	binary.BigEndian.PutUint32(chunk[16:], crc32.ChecksumIEEE(chunk[4:16]))
	apng := append([]byte{}, pngData.Bytes()[:33]...)
	apng = append(apng, chunk...)
	apng = append(apng, pngData.Bytes()[33:]...)
	if _, err := (Decoder{}).Decode(apng); !errors.Is(err, imageprocessing.ErrUnsupportedVariant) {
		t.Fatalf("APNG: %v", err)
	}
	// Minimal TIFF directory that links to another page must never be flattened.
	tiffData := []byte{'I', 'I', 42, 0, 8, 0, 0, 0, 0, 0, 14, 0, 0, 0, 0, 0, 0, 0}
	if _, err := (Decoder{}).Decode(tiffData); !errors.Is(err, imageprocessing.ErrUnsupportedVariant) {
		t.Fatalf("multi-page TIFF: %v", err)
	}
}

func TestResizeRejectsOversizedHeadersBeforeDecoding(t *testing.T) {
	var raw bytes.Buffer
	if err := png.Encode(&raw, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	data := raw.Bytes()
	binary.BigEndian.PutUint32(data[16:20], 8001)
	binary.BigEndian.PutUint32(data[20:24], 4000)
	binary.BigEndian.PutUint32(data[29:33], crc32.ChecksumIEEE(data[12:29]))
	if _, err := (Decoder{}).Decode(data); !errors.Is(err, imageprocessing.ErrImageTooLarge) {
		t.Fatalf("PNG input pixel budget: %v", err)
	}

	raw.Reset()
	encoder := animation.NewEncoder(&raw, 2, 2, &animation.EncodeOptions{Lossless: true})
	for _, c := range []color.RGBA{{R: 255, A: 255}, {G: 255, A: 255}, {B: 255, A: 255}} {
		frame := image.NewRGBA(image.Rect(0, 0, 2, 2))
		for y := 0; y < 2; y++ {
			for x := 0; x < 2; x++ {
				frame.Set(x, y, c)
			}
		}
		if err := encoder.AddFrame(frame, 100*time.Millisecond); err != nil {
			t.Fatal(err)
		}
	}
	if err := encoder.Close(); err != nil {
		t.Fatal(err)
	}
	data = raw.Bytes()
	features, err := webp.GetFeatures(bytes.NewReader(data))
	if err != nil || features.FrameCount != 3 {
		t.Fatalf("invalid fixture: %+v %v", features, err)
	}
	pos := bytes.Index(data, []byte("VP8X"))
	if pos < 0 {
		t.Fatal("missing extended header")
	}
	// VP8X stores canvas dimensions minus one, as little-endian 24-bit values.
	for i := 0; i < 3; i++ {
		data[pos+12+i] = byte(uint32(7999) >> (8 * i))
		data[pos+15+i] = byte(uint32(3999) >> (8 * i))
	}
	if _, err := (Decoder{}).Decode(data); !errors.Is(err, imageprocessing.ErrAnimationTooLarge) {
		t.Fatalf("WebP input frame budget: %v", err)
	}
}
