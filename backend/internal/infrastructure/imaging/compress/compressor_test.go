package compress

import (
	"bytes"
	"encoding/binary"
	"errors"
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
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/imaging"
	"github.com/strukturag/libheif/go/heif"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

func TestCompressorCompressesSupportedStaticFormats(t *testing.T) {
	source := gradientImage(40, 28)

	tests := []struct {
		name        string
		input       []byte
		format      imagecompression.Format
		contentType string
		decode      func(*testing.T, []byte) image.Image
	}{
		{
			name:        "jpeg",
			input:       encodeJPEG(t, source, 100),
			format:      imagecompression.FormatJPEG,
			contentType: "image/jpeg",
			decode:      decodeJPEG,
		},
		{
			name:        "png",
			input:       encodePNG(t, source, png.NoCompression),
			format:      imagecompression.FormatPNG,
			contentType: "image/png",
			decode:      decodePNG,
		},
		{
			name:        "webp",
			input:       encodeWebP(t, source, &webp.EncoderOptions{Quality: 100, Method: 4}),
			format:      imagecompression.FormatWebP,
			contentType: "image/webp",
			decode:      decodeWebP,
		},
		{
			name:        "avif",
			input:       encodeAVIF(t, source),
			format:      imagecompression.FormatAVIF,
			contentType: "image/avif",
			decode:      decodeAVIF,
		},
		{
			name:        "heic",
			input:       encodeHEIF(t, source),
			format:      imagecompression.FormatHEIF,
			contentType: "image/heic",
			decode:      decodeHEIF,
		},
		{
			name:        "gif",
			input:       encodeStaticGIF(t, source),
			format:      imagecompression.FormatGIF,
			contentType: "image/gif",
			decode:      decodeGIF,
		},
		{
			name:        "bmp",
			input:       encodeBMP(t, source),
			format:      imagecompression.FormatBMP,
			contentType: "image/bmp",
			decode:      decodeBMP,
		},
		{
			name:        "tiff",
			input:       encodeTIFF(t, source),
			format:      imagecompression.FormatTIFF,
			contentType: "image/tiff",
			decode:      decodeTIFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewCompressor().Compress(tt.input)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if result.Format != tt.format {
				t.Fatalf("expected format %q, got %q", tt.format, result.Format)
			}
			if result.ContentType != tt.contentType {
				t.Fatalf("expected content type %q, got %q", tt.contentType, result.ContentType)
			}
			if result.Animated {
				t.Fatalf("expected static result")
			}
			if result.FrameCount != 1 {
				t.Fatalf("expected frame count 1, got %d", result.FrameCount)
			}
			if result.Width != 40 || result.Height != 28 {
				t.Fatalf("expected result dimensions 40x28, got %dx%d", result.Width, result.Height)
			}
			if len(result.Data) == 0 {
				t.Fatal("expected compressed output bytes")
			}
			assertDetectedFormat(t, result.Data, tt.format, tt.contentType)

			decoded := tt.decode(t, result.Data)
			assertDimensions(t, decoded, 40, 28)
		})
	}
}

func TestCompressorReducesDeterministicHighQualityJPEGFixture(t *testing.T) {
	source := detailedImage(120, 90)
	input := encodeJPEG(t, source, 100)

	result, err := NewCompressor().Compress(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result.Data) >= len(input) {
		t.Fatalf("expected high-quality JPEG fixture to shrink, original=%d optimized=%d", len(input), len(result.Data))
	}
}

func TestCompressorPreservesPNGTransparencyLosslessly(t *testing.T) {
	source := alphaImage(24, 18)
	input := encodePNG(t, source, png.NoCompression)

	result, err := NewCompressor().Compress(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	decoded := decodePNG(t, result.Data)
	assertPixelsEqual(t, source, decoded)
}

func TestCompressorPreservesLosslessWebPTransparency(t *testing.T) {
	source := alphaImage(18, 14)
	input := encodeWebP(t, source, &webp.EncoderOptions{
		Lossless: true,
		Quality:  100,
		Method:   6,
		Exact:    true,
	})

	result, err := NewCompressor().Compress(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	decoded := decodeWebP(t, result.Data)
	assertPixelsEqual(t, source, decoded)
}

func TestCompressorPreservesGIFAnimationFramesAndTiming(t *testing.T) {
	input := encodeAnimatedGIF(t, 16, 12)

	result, err := NewCompressor().Compress(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Format != imagecompression.FormatGIF || result.ContentType != "image/gif" {
		t.Fatalf("expected GIF result, got %q %q", result.Format, result.ContentType)
	}
	if len(result.Data) == 0 {
		t.Fatal("expected compressed GIF bytes")
	}
	assertDetectedFormat(t, result.Data, imagecompression.FormatGIF, "image/gif")
	if !result.Animated || result.FrameCount != 2 {
		t.Fatalf("expected two-frame animation, animated=%v frameCount=%d", result.Animated, result.FrameCount)
	}

	decoded, err := gif.DecodeAll(bytes.NewReader(result.Data))
	if err != nil {
		t.Fatalf("compressed GIF could not be decoded: %v", err)
	}
	if len(decoded.Image) != 2 {
		t.Fatalf("expected 2 GIF frames, got %d", len(decoded.Image))
	}
	if decoded.Delay[0] != 10 || decoded.Delay[1] != 24 {
		t.Fatalf("expected GIF delays [10 24], got %v", decoded.Delay)
	}
	if decoded.LoopCount != 0 {
		t.Fatalf("expected infinite GIF loop count, got %d", decoded.LoopCount)
	}
	if len(decoded.Disposal) != 2 || decoded.Disposal[0] != gif.DisposalNone || decoded.Disposal[1] != gif.DisposalBackground {
		t.Fatalf("expected GIF disposal [none background], got %v", decoded.Disposal)
	}
}

func TestCompressorPreservesAnimatedWebPFramesAndTiming(t *testing.T) {
	input := encodeAnimatedWebP(t, 18, 10)

	result, err := NewCompressor().Compress(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Format != imagecompression.FormatWebP || result.ContentType != "image/webp" {
		t.Fatalf("expected WebP result, got %q %q", result.Format, result.ContentType)
	}
	if len(result.Data) == 0 {
		t.Fatal("expected compressed WebP bytes")
	}
	assertDetectedFormat(t, result.Data, imagecompression.FormatWebP, "image/webp")
	if !result.Animated || result.FrameCount != 2 {
		t.Fatalf("expected two-frame animation, animated=%v frameCount=%d", result.Animated, result.FrameCount)
	}

	features, err := webp.GetFeatures(bytes.NewReader(result.Data))
	if err != nil {
		t.Fatalf("compressed WebP features could not be read: %v", err)
	}
	if !features.HasAnimation || features.FrameCount != 2 {
		t.Fatalf("expected animated WebP with 2 frames, got animated=%v frames=%d", features.HasAnimation, features.FrameCount)
	}

	anim, err := animation.Decode(bytes.NewReader(result.Data))
	if err != nil {
		t.Fatalf("compressed WebP animation could not be decoded: %v", err)
	}
	if len(anim.Frames) != 2 {
		t.Fatalf("expected 2 decoded WebP frames, got %d", len(anim.Frames))
	}
	if anim.Frames[0].Duration != 100*time.Millisecond || anim.Frames[1].Duration != 240*time.Millisecond {
		t.Fatalf("unexpected WebP frame durations: %v and %v", anim.Frames[0].Duration, anim.Frames[1].Duration)
	}
}

func TestCompressorNormalizesJPEGEXIFOrientation(t *testing.T) {
	source := gradientImage(9, 4)
	input := withEXIFOrientation(t, encodeJPEG(t, source, 100), 6)

	result, err := NewCompressor().Compress(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Width != 4 || result.Height != 9 {
		t.Fatalf("expected orientation-normalized dimensions 4x9, got %dx%d", result.Width, result.Height)
	}

	decoded := decodeJPEG(t, result.Data)
	assertDimensions(t, decoded, 4, 9)
}

func TestCompressorRejectsUnsupportedInput(t *testing.T) {
	_, err := NewCompressor().Compress([]byte("plain text"))

	if !errors.Is(err, imagecompression.ErrUnsupportedFormat) {
		t.Fatalf("expected ErrUnsupportedFormat, got %v", err)
	}
}

func TestCompressorRejectsUnsupportedCameraRAWInput(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{name: "dng", input: minimalDNG()},
		{name: "cr2", input: []byte{'I', 'I', '*', 0, 0x10, 0, 0, 0, 'C', 'R', 0x02, 0}},
		{name: "cr3", input: corruptBMFF("crx ")},
		{name: "raf", input: []byte("FUJIFILMCCD-RAW 0201")},
		{name: "rw2", input: []byte{'I', 'I', 'U', 0, 0x18, 0, 0, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCompressor().Compress(tt.input)
			if !errors.Is(err, imagecompression.ErrUnsupportedFormat) {
				t.Fatalf("expected ErrUnsupportedFormat, got %v", err)
			}
		})
	}
}

func TestCompressorRejectsCorruptedImageContentForSupportedFormats(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{name: "jpeg", input: []byte{0xff, 0xd8, 0xff, 0x00}},
		{name: "png", input: []byte("\x89PNG\r\n\x1a\nnot a png")},
		{name: "webp", input: []byte("RIFF\x04\x00\x00\x00WEBP")},
		{name: "avif", input: corruptBMFF("avif")},
		{name: "heic", input: corruptBMFF("heic")},
		{name: "gif", input: []byte{'G', 'I', 'F', '8', '9', 'a', 1, 0, 1, 0, 0, 0, 0, ';'}},
		{name: "bmp", input: []byte("BMnot a bmp")},
		{name: "tiff", input: []byte("II*\x00not a tiff")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCompressor().Compress(tt.input)
			if !errors.Is(err, imagecompression.ErrInvalidImage) {
				t.Fatalf("expected ErrInvalidImage, got %v", err)
			}
		})
	}
}

func TestCompressorRejectsImagesAboveDecodedPixelLimitForEveryFormat(t *testing.T) {
	source := gradientImage(2, 2)

	tests := []struct {
		name  string
		input []byte
	}{
		{name: "jpeg", input: encodeJPEG(t, source, 100)},
		{name: "png", input: encodePNG(t, source, png.DefaultCompression)},
		{name: "webp", input: encodeWebP(t, source, &webp.EncoderOptions{Quality: 90})},
		{name: "avif", input: encodeAVIF(t, source)},
		{name: "heic", input: encodeHEIF(t, source)},
		{name: "gif", input: encodeStaticGIF(t, source)},
		{name: "bmp", input: encodeBMP(t, source)},
		{name: "tiff", input: encodeTIFF(t, source)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressor := Compressor{
				JPEGQuality:      DefaultJPEGQuality,
				WebPQuality:      DefaultWebPQuality,
				AVIFQuality:      DefaultAVIFQuality,
				HEIFQuality:      DefaultHEIFQuality,
				MaxDecodedPixels: 3,
			}

			_, err := compressor.Compress(tt.input)
			if !errors.Is(err, imagecompression.ErrImageTooLarge) {
				t.Fatalf("expected ErrImageTooLarge, got %v", err)
			}
		})
	}
}

func TestCompressorRejectsAnimationsAboveFramePixelLimit(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{name: "gif", input: encodeAnimatedGIF(t, 3, 2)},
		{name: "webp", input: encodeAnimatedWebP(t, 3, 2)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressor := NewCompressor()
			compressor.MaxAnimatedFramePixels = 11

			_, err := compressor.Compress(tt.input)
			if !errors.Is(err, imagecompression.ErrAnimationTooLarge) {
				t.Fatalf("expected ErrAnimationTooLarge, got %v", err)
			}
		})
	}
}

func gradientImage(width, height int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8((x * 255) / max(1, width-1)),
				G: uint8((y * 255) / max(1, height-1)),
				B: uint8((x*y + x + y) % 256),
				A: 255,
			})
		}
	}

	return img
}

func detailedImage(width, height int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			v := (x*37 + y*53 + x*y*11) % 256
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8(v),
				G: uint8((v + x*7) % 256),
				B: uint8((v + y*13) % 256),
				A: 255,
			})
		}
	}

	return img
}

func alphaImage(width, height int) *image.NRGBA {
	img := gradientImage(width, height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := img.NRGBAAt(x, y)
			if (x+y)%3 == 0 {
				pixel.A = 0
			} else if (x+y)%3 == 1 {
				pixel.A = 128
			}
			img.SetNRGBA(x, y, pixel)
		}
	}

	return img
}

func encodeJPEG(t *testing.T, img image.Image, quality int) []byte {
	t.Helper()

	var buffer bytes.Buffer
	if err := jpeg.Encode(&buffer, img, &jpeg.Options{Quality: quality}); err != nil {
		t.Fatalf("failed to encode jpeg fixture: %v", err)
	}

	return buffer.Bytes()
}

func encodePNG(t *testing.T, img image.Image, level png.CompressionLevel) []byte {
	t.Helper()

	var buffer bytes.Buffer
	encoder := png.Encoder{CompressionLevel: level}
	if err := encoder.Encode(&buffer, img); err != nil {
		t.Fatalf("failed to encode png fixture: %v", err)
	}

	return buffer.Bytes()
}

func encodeWebP(t *testing.T, img image.Image, options *webp.EncoderOptions) []byte {
	t.Helper()

	var buffer bytes.Buffer
	if err := webp.Encode(&buffer, img, options); err != nil {
		t.Fatalf("failed to encode webp fixture: %v", err)
	}

	return buffer.Bytes()
}

func encodeAVIF(t *testing.T, img image.Image) []byte {
	t.Helper()

	var buffer bytes.Buffer
	if err := avif.Encode(&buffer, img, avif.Options{Quality: 70, QualityAlpha: 100, Speed: 8}); err != nil {
		t.Fatalf("failed to encode avif fixture: %v", err)
	}

	return buffer.Bytes()
}

func encodeHEIF(t *testing.T, img image.Image) []byte {
	t.Helper()

	ctx, err := heif.EncodeFromImage(toRGBA(img), heif.CompressionHEVC, 70, heif.LosslessModeDisabled, heif.LoggingLevelNone)
	if err != nil {
		t.Fatalf("failed to encode heif fixture: %v", err)
	}

	path := filepath.Join(t.TempDir(), "fixture.heic")
	if err := ctx.WriteToFile(path); err != nil {
		t.Fatalf("failed to write heif fixture: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read heif fixture: %v", err)
	}

	return data
}

func encodeStaticGIF(t *testing.T, img image.Image) []byte {
	t.Helper()

	palette := color.Palette{
		color.Black,
		color.White,
		color.NRGBA{R: 64, G: 160, B: 220, A: 255},
		color.NRGBA{R: 230, G: 80, B: 110, A: 255},
	}
	bounds := img.Bounds()
	frame := image.NewPaletted(image.Rect(0, 0, bounds.Dx(), bounds.Dy()), palette)
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			frame.Set(x, y, img.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}

	var buffer bytes.Buffer
	if err := gif.EncodeAll(&buffer, &gif.GIF{
		Image: []*image.Paletted{frame},
		Delay: []int{0},
		Config: image.Config{
			ColorModel: palette,
			Width:      bounds.Dx(),
			Height:     bounds.Dy(),
		},
	}); err != nil {
		t.Fatalf("failed to encode gif fixture: %v", err)
	}

	return buffer.Bytes()
}

func encodeAnimatedGIF(t *testing.T, width, height int) []byte {
	t.Helper()

	palette := color.Palette{color.Black, color.White, color.RGBA{R: 255, A: 255}, color.RGBA{B: 255, A: 255}}
	first := image.NewPaletted(image.Rect(0, 0, width, height), palette)
	second := image.NewPaletted(image.Rect(0, 0, width, height), palette)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			first.SetColorIndex(x, y, uint8((x+y)%len(palette)))
			second.SetColorIndex(x, y, uint8((x*2+y+1)%len(palette)))
		}
	}

	var buffer bytes.Buffer
	if err := gif.EncodeAll(&buffer, &gif.GIF{
		Image:     []*image.Paletted{first, second},
		Delay:     []int{10, 24},
		Disposal:  []byte{gif.DisposalNone, gif.DisposalBackground},
		LoopCount: 0,
		Config: image.Config{
			ColorModel: palette,
			Width:      width,
			Height:     height,
		},
	}); err != nil {
		t.Fatalf("failed to encode animated gif fixture: %v", err)
	}

	return buffer.Bytes()
}

func encodeAnimatedWebP(t *testing.T, width, height int) []byte {
	t.Helper()

	first := gradientImage(width, height)
	second := detailedImage(width, height)

	var buffer bytes.Buffer
	encoder := animation.NewEncoder(&buffer, width, height, &animation.EncodeOptions{
		LoopCount: 0,
		Quality:   85,
		Lossless:  true,
	})
	if err := encoder.AddFrame(first, 100*time.Millisecond); err != nil {
		t.Fatalf("failed to add first webp frame: %v", err)
	}
	if err := encoder.AddFrame(second, 240*time.Millisecond); err != nil {
		t.Fatalf("failed to add second webp frame: %v", err)
	}
	if err := encoder.Close(); err != nil {
		t.Fatalf("failed to encode animated webp fixture: %v", err)
	}

	return buffer.Bytes()
}

func encodeBMP(t *testing.T, img image.Image) []byte {
	t.Helper()

	var buffer bytes.Buffer
	if err := bmp.Encode(&buffer, img); err != nil {
		t.Fatalf("failed to encode bmp fixture: %v", err)
	}

	return buffer.Bytes()
}

func encodeTIFF(t *testing.T, img image.Image) []byte {
	t.Helper()

	var buffer bytes.Buffer
	if err := tiff.Encode(&buffer, img, &tiff.Options{Compression: tiff.Uncompressed}); err != nil {
		t.Fatalf("failed to encode tiff fixture: %v", err)
	}

	return buffer.Bytes()
}

func decodeJPEG(t *testing.T, data []byte) image.Image {
	t.Helper()

	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to decode jpeg: %v", err)
	}

	return img
}

func decodePNG(t *testing.T, data []byte) image.Image {
	t.Helper()

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to decode png: %v", err)
	}

	return img
}

func decodeWebP(t *testing.T, data []byte) image.Image {
	t.Helper()

	img, err := webp.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to decode webp: %v", err)
	}

	return img
}

func decodeAVIF(t *testing.T, data []byte) image.Image {
	t.Helper()

	img, err := avif.Decode(bytes.NewReader(data), avif.Options{AutoRotate: true})
	if err != nil {
		t.Fatalf("failed to decode avif: %v", err)
	}

	return img
}

func decodeHEIF(t *testing.T, data []byte) image.Image {
	t.Helper()

	ctx, err := heif.NewContext()
	if err != nil {
		t.Fatalf("failed to create heif context: %v", err)
	}
	if err := ctx.ReadFromMemory(data); err != nil {
		t.Fatalf("failed to read heif data: %v", err)
	}
	handle, err := ctx.GetPrimaryImageHandle()
	if err != nil {
		t.Fatalf("failed to get heif primary handle: %v", err)
	}
	decoded, err := handle.DecodeImage(heif.ColorspaceRGB, heif.ChromaInterleavedRGBA, nil)
	if err != nil {
		t.Fatalf("failed to decode heif: %v", err)
	}
	img, err := decoded.GetImage()
	if err != nil {
		t.Fatalf("failed to convert heif image: %v", err)
	}

	return img
}

func decodeGIF(t *testing.T, data []byte) image.Image {
	t.Helper()

	img, err := gif.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to decode gif: %v", err)
	}

	return img
}

func decodeBMP(t *testing.T, data []byte) image.Image {
	t.Helper()

	img, err := bmp.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to decode bmp: %v", err)
	}

	return img
}

func decodeTIFF(t *testing.T, data []byte) image.Image {
	t.Helper()

	img, err := tiff.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to decode tiff: %v", err)
	}

	return img
}

func withEXIFOrientation(t *testing.T, jpegData []byte, orientation uint16) []byte {
	t.Helper()

	if len(jpegData) < 2 || jpegData[0] != 0xff || jpegData[1] != 0xd8 {
		t.Fatal("fixture is not a jpeg")
	}

	var payload bytes.Buffer
	payload.WriteString("Exif\x00\x00")
	payload.Write([]byte{'M', 'M', 0x00, 0x2a})
	_ = binary.Write(&payload, binary.BigEndian, uint32(8))
	_ = binary.Write(&payload, binary.BigEndian, uint16(1))
	_ = binary.Write(&payload, binary.BigEndian, uint16(0x0112))
	_ = binary.Write(&payload, binary.BigEndian, uint16(3))
	_ = binary.Write(&payload, binary.BigEndian, uint32(1))
	_ = binary.Write(&payload, binary.BigEndian, orientation)
	_ = binary.Write(&payload, binary.BigEndian, uint16(0))
	_ = binary.Write(&payload, binary.BigEndian, uint32(0))

	app1Length := payload.Len() + 2
	if app1Length > 0xffff {
		t.Fatalf("EXIF payload too large: %d", app1Length)
	}

	output := make([]byte, 0, len(jpegData)+payload.Len()+4)
	output = append(output, jpegData[:2]...)
	output = append(output, 0xff, 0xe1, byte(app1Length>>8), byte(app1Length))
	output = append(output, payload.Bytes()...)
	output = append(output, jpegData[2:]...)

	return output
}

func corruptBMFF(brand string) []byte {
	data := make([]byte, 24)
	binary.BigEndian.PutUint32(data[:4], uint32(len(data)))
	copy(data[4:8], "ftyp")
	copy(data[8:12], brand)
	copy(data[16:20], brand)
	copy(data[20:24], "junk")

	return data
}

func minimalDNG() []byte {
	var data bytes.Buffer

	data.Write([]byte{'I', 'I', '*', 0})
	_ = binary.Write(&data, binary.LittleEndian, uint32(8))
	_ = binary.Write(&data, binary.LittleEndian, uint16(1))
	_ = binary.Write(&data, binary.LittleEndian, uint16(0xc612))
	_ = binary.Write(&data, binary.LittleEndian, uint16(1))
	_ = binary.Write(&data, binary.LittleEndian, uint32(4))
	data.Write([]byte{1, 4, 0, 0})
	_ = binary.Write(&data, binary.LittleEndian, uint32(0))

	return data.Bytes()
}

func assertDimensions(t *testing.T, img image.Image, width, height int) {
	t.Helper()

	bounds := img.Bounds()
	if bounds.Dx() != width || bounds.Dy() != height {
		t.Fatalf("expected dimensions %dx%d, got %dx%d", width, height, bounds.Dx(), bounds.Dy())
	}
}

func assertDetectedFormat(t *testing.T, data []byte, format imagecompression.Format, contentType string) {
	t.Helper()

	detected, err := imaging.DetectFormat(data)
	if err != nil {
		t.Fatalf("compressed output format could not be detected: %v", err)
	}
	if detected.Format != format {
		t.Fatalf("expected detected format %q, got %q", format, detected.Format)
	}
	if detected.ContentType != contentType {
		t.Fatalf("expected detected content type %q, got %q", contentType, detected.ContentType)
	}
}

func assertPixelsEqual(t *testing.T, expected image.Image, actual image.Image) {
	t.Helper()

	expectedBounds := expected.Bounds()
	if expectedBounds.Dx() != actual.Bounds().Dx() || expectedBounds.Dy() != actual.Bounds().Dy() {
		t.Fatalf("image bounds differ: expected %v, got %v", expectedBounds, actual.Bounds())
	}

	for y := expectedBounds.Min.Y; y < expectedBounds.Max.Y; y++ {
		for x := expectedBounds.Min.X; x < expectedBounds.Max.X; x++ {
			expectedColor := color.NRGBAModel.Convert(expected.At(x, y)).(color.NRGBA)
			actualColor := color.NRGBAModel.Convert(actual.At(x, y)).(color.NRGBA)

			if expectedColor != actualColor {
				t.Fatalf("pixel mismatch at %d,%d: expected %#v, got %#v", x, y, expectedColor, actualColor)
			}
		}
	}
}
