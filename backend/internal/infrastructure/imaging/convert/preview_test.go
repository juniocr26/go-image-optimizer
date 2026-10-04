package convert

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageprocessing"
	"golang.org/x/image/tiff"
	"hash/crc32"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"testing"
)

func TestPreviewFixtures(t *testing.T) {
	for _, ext := range []string{"jpg", "png", "webp", "avif", "heic", "heif", "gif", "bmp", "tiff"} {
		t.Run(ext, func(t *testing.T) {
			root := os.Getenv("TESTDATA_DIR")
			if root == "" {
				root = "/testdata/images"
			}
			input, err := os.ReadFile(root + "/sample." + ext)
			if err != nil {
				t.Fatal(err)
			}
			before := bytes.Clone(input)
			result, err := (Processor{}).Preview(context.Background(), input)
			if err != nil {
				t.Fatal(err)
			}
			decoded, format, err := image.Decode(bytes.NewReader(result.Data))
			if err != nil {
				t.Fatal(err)
			}
			if result.ContentType != "image/"+format {
				t.Fatal(result.ContentType, format)
			}
			if decoded.Bounds().Dx() > 1200 || decoded.Bounds().Dy() > 1200 {
				t.Fatal(decoded.Bounds())
			}
			info, err := (Processor{}).Inspect(context.Background(), input)
			if err != nil {
				t.Fatal(err)
			}
			if abs(result.Width*info.Height-result.Height*info.Width) > max(info.Width, info.Height) {
				t.Fatal("aspect ratio changed")
			}
			if result.Width > info.Width || result.Height > info.Height {
				t.Fatal("enlarged")
			}
			if !bytes.Equal(input, before) {
				t.Fatal("source modified")
			}
		})
	}
}

func TestPreviewBoundsAlphaAndErrors(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2400, 600))
	img.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 128})
	var input bytes.Buffer
	if err := png.Encode(&input, img); err != nil {
		t.Fatal(err)
	}
	result, err := (Processor{}).Preview(context.Background(), input.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if result.Width != 1200 || result.Height != 300 || result.ContentType != "image/png" {
		t.Fatal(result)
	}
	decoded, err := png.Decode(bytes.NewReader(result.Data))
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, alpha := decoded.At(1199, 299).RGBA()
	if alpha != 0 {
		t.Fatal("lost alpha")
	}
	if _, err := (Processor{}).Preview(context.Background(), []byte("invalid")); err == nil {
		t.Fatal("accepted invalid")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := (Processor{}).Preview(ctx, input.Bytes()); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	huge := bytes.Clone(input.Bytes())
	huge[16], huge[17], huge[18], huge[19] = 0, 1, 0, 0
	if _, err := (Processor{}).Preview(context.Background(), huge); err == nil {
		t.Fatal("accepted invalid dimensions")
	}
	if _, err := (Processor{}).Preview(context.Background(), nil); !errors.Is(err, imageprocessing.ErrEmptyImage) {
		t.Fatal(err)
	}
}

func TestPreviewOrientationVariantsAndLimits(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 40, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 40; x++ {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	var raw bytes.Buffer
	jpeg.Encode(&raw, img, nil)
	payload := []byte{'E', 'x', 'i', 'f', 0, 0, 'I', 'I', 42, 0, 8, 0, 0, 0, 1, 0, 0x12, 1, 3, 0, 1, 0, 0, 0, 6, 0, 0, 0, 0, 0, 0, 0}
	input := append([]byte{}, raw.Bytes()[:2]...)
	input = append(input, 0xff, 0xe1, 0, byte(len(payload)+2))
	input = append(input, payload...)
	input = append(input, raw.Bytes()[2:]...)
	result, err := (Processor{}).Preview(context.Background(), input)
	if err != nil || result.Width != 20 || result.Height != 40 || result.ContentType != "image/jpeg" {
		t.Fatal(result, err)
	}
	frame := image.NewPaletted(image.Rect(0, 0, 2, 2), color.Palette{color.Black, color.White})
	raw.Reset()
	gif.EncodeAll(&raw, &gif.GIF{Image: []*image.Paletted{frame, frame}, Delay: []int{1, 1}})
	if _, err := (Processor{}).Preview(context.Background(), raw.Bytes()); !errors.Is(err, imageprocessing.ErrUnsupportedVariant) {
		t.Fatal(err)
	}
	raw.Reset()
	tiff.Encode(&raw, img, nil)
	multi := bytes.Clone(raw.Bytes())
	offset := binary.LittleEndian.Uint32(multi[4:8])
	count := binary.LittleEndian.Uint16(multi[offset : offset+2])
	next := int(offset) + 2 + 12*int(count)
	binary.LittleEndian.PutUint32(multi[next:next+4], offset)
	if _, err := (Processor{}).Preview(context.Background(), multi); !errors.Is(err, imageprocessing.ErrUnsupportedVariant) {
		t.Fatal(err)
	}
	huge := fixture(t)
	binary.BigEndian.PutUint32(huge[16:20], 100000)
	binary.BigEndian.PutUint32(huge[20:24], 100000)
	binary.BigEndian.PutUint32(huge[29:33], crc32.ChecksumIEEE(huge[12:29]))
	if _, err := (Processor{}).Preview(context.Background(), huge); !errors.Is(err, imageprocessing.ErrImageTooLarge) {
		t.Fatal(err)
	}

	data := fixture(t)
	chunk := make([]byte, 20)
	binary.BigEndian.PutUint32(chunk[:4], 8)
	copy(chunk[4:8], "acTL")
	binary.BigEndian.PutUint32(chunk[8:12], 2)
	binary.BigEndian.PutUint32(chunk[16:], crc32.ChecksumIEEE(chunk[4:16]))
	apng := append(bytes.Clone(data[:33]), chunk...)
	apng = append(apng, data[33:]...)
	if _, err := (Processor{}).Preview(context.Background(), apng); !errors.Is(err, imageprocessing.ErrUnsupportedVariant) {
		t.Fatal(err)
	}
	raw.Reset()
	gif.EncodeAll(&raw, &gif.GIF{Image: []*image.Paletted{frame, frame, frame}, Delay: []int{1, 1, 1}})
	bomb := raw.Bytes()
	binary.LittleEndian.PutUint16(bomb[6:8], 8000)
	binary.LittleEndian.PutUint16(bomb[8:10], 4000)
	if _, err := (Processor{}).Preview(context.Background(), bomb); !errors.Is(err, imageprocessing.ErrAnimationTooLarge) {
		t.Fatal(err)
	}
	if _, err := (Processor{}).Preview(context.Background(), make([]byte, (50<<20)+1)); !errors.Is(err, imageprocessing.ErrImageTooLarge) {
		t.Fatal(err)
	}
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
