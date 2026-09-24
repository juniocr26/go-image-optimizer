package compress

import (
	"bytes"
	"image"
	"image/gif"
	"os"
	"path/filepath"
	"testing"

	"github.com/deepteams/webp"
	"github.com/deepteams/webp/animation"
	"github.com/gen2brain/avif"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
)

func TestCompressorRealFixtures(t *testing.T) {
	fixturesDir := realImageFixturesDir(t)

	tests := []struct {
		name        string
		filename    string
		format      imagecompression.Format
		contentType string
		decode      func(*testing.T, []byte) image.Image
		validate    func(*testing.T, []byte, imagecompression.Result)
	}{
		{
			name:        "jpeg",
			filename:    "sample.jpg",
			format:      imagecompression.FormatJPEG,
			contentType: "image/jpeg",
			decode:      decodeJPEG,
		},
		{
			name:        "png",
			filename:    "sample.png",
			format:      imagecompression.FormatPNG,
			contentType: "image/png",
			decode:      decodePNG,
			validate:    assertRealPNGFixture,
		},
		{
			name:        "webp",
			filename:    "sample.webp",
			format:      imagecompression.FormatWebP,
			contentType: "image/webp",
			decode:      decodeWebP,
			validate:    assertRealWebPFixture,
		},
		{
			name:        "avif",
			filename:    "sample.avif",
			format:      imagecompression.FormatAVIF,
			contentType: "image/avif",
			decode:      decodeAVIF,
			validate:    assertRealAVIFFixture,
		},
		{
			name:        "heic",
			filename:    "sample.heic",
			format:      imagecompression.FormatHEIF,
			contentType: "image/heic",
			decode:      decodeHEIF,
		},
		{
			name:        "heif",
			filename:    "sample.heif",
			format:      imagecompression.FormatHEIF,
			contentType: "image/heic",
			decode:      decodeHEIF,
		},
		{
			name:        "gif",
			filename:    "sample.gif",
			format:      imagecompression.FormatGIF,
			contentType: "image/gif",
			decode:      decodeGIF,
			validate:    assertRealGIFFixture,
		},
		{
			name:        "bmp",
			filename:    "sample.bmp",
			format:      imagecompression.FormatBMP,
			contentType: "image/bmp",
			decode:      decodeBMP,
		},
		{
			name:        "tiff",
			filename:    "sample.tiff",
			format:      imagecompression.FormatTIFF,
			contentType: "image/tiff",
			decode:      decodeTIFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := readRealImageFixture(t, fixturesDir, tt.filename)
			assertDetectedFormat(t, source, tt.format, tt.contentType)

			sourceImage := tt.decode(t, source)
			width, height := dimensions(sourceImage)

			result, err := NewCompressor().Compress(source)
			if err != nil {
				t.Fatalf("expected real fixture %s to compress successfully: %v", tt.filename, err)
			}

			if len(result.Data) == 0 {
				t.Fatalf("expected compressed output bytes for %s", tt.filename)
			}
			if result.Format != tt.format {
				t.Fatalf("expected result format %q for %s, got %q", tt.format, tt.filename, result.Format)
			}
			if result.ContentType != tt.contentType {
				t.Fatalf("expected result content type %q for %s, got %q", tt.contentType, tt.filename, result.ContentType)
			}
			if result.Width != width || result.Height != height {
				t.Fatalf("expected result dimensions %dx%d for %s, got %dx%d", width, height, tt.filename, result.Width, result.Height)
			}

			assertDetectedFormat(t, result.Data, tt.format, tt.contentType)
			assertDimensions(t, tt.decode(t, result.Data), width, height)

			if tt.validate != nil {
				tt.validate(t, source, result)
			}
		})
	}
}

func realImageFixturesDir(t *testing.T) string {
	t.Helper()

	if dir := os.Getenv("REAL_IMAGE_FIXTURES_DIR"); dir != "" {
		return dir
	}

	candidates := []string{
		filepath.Join("..", "..", "..", "..", "storage", "testdata", "images"),
		filepath.Join("testdata", "images"),
	}

	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return candidate
		}
	}

	t.Fatal("real image fixture directory not found; set REAL_IMAGE_FIXTURES_DIR or run tests from the repository checkout")
	return ""
}

func readRealImageFixture(t *testing.T, fixturesDir, filename string) []byte {
	t.Helper()

	path := filepath.Join(fixturesDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read real image fixture %s: %v", path, err)
	}
	if len(data) == 0 {
		t.Fatalf("real image fixture %s is empty", path)
	}

	return data
}

func dimensions(img image.Image) (int, int) {
	bounds := img.Bounds()
	return bounds.Dx(), bounds.Dy()
}

func assertRealPNGFixture(t *testing.T, source []byte, result imagecompression.Result) {
	t.Helper()

	assertPixelsEqual(t, decodePNG(t, source), decodePNG(t, result.Data))
}

func assertRealWebPFixture(t *testing.T, source []byte, result imagecompression.Result) {
	t.Helper()

	sourceFeatures, err := webp.GetFeatures(bytes.NewReader(source))
	if err != nil {
		t.Fatalf("failed to read source WebP features: %v", err)
	}
	resultFeatures, err := webp.GetFeatures(bytes.NewReader(result.Data))
	if err != nil {
		t.Fatalf("failed to read compressed WebP features: %v", err)
	}

	sourceFrameCount := max(1, sourceFeatures.FrameCount)
	if result.FrameCount != sourceFrameCount {
		t.Fatalf("expected WebP frame count %d, got %d", sourceFrameCount, result.FrameCount)
	}
	if result.Animated != sourceFeatures.HasAnimation {
		t.Fatalf("expected WebP animated=%v, got %v", sourceFeatures.HasAnimation, result.Animated)
	}
	if sourceFeatures.HasAlpha && !resultFeatures.HasAlpha {
		t.Fatal("expected compressed WebP to preserve alpha support")
	}

	if !sourceFeatures.HasAnimation {
		return
	}

	sourceAnimation, err := animation.Decode(bytes.NewReader(source))
	if err != nil {
		t.Fatalf("failed to decode source WebP animation: %v", err)
	}
	resultAnimation, err := animation.Decode(bytes.NewReader(result.Data))
	if err != nil {
		t.Fatalf("failed to decode compressed WebP animation: %v", err)
	}
	if len(resultAnimation.Frames) != len(sourceAnimation.Frames) {
		t.Fatalf("expected %d WebP frames, got %d", len(sourceAnimation.Frames), len(resultAnimation.Frames))
	}
	for i := range sourceAnimation.Frames {
		if resultAnimation.Frames[i].Duration != sourceAnimation.Frames[i].Duration {
			t.Fatalf("expected WebP frame %d duration %s, got %s", i, sourceAnimation.Frames[i].Duration, resultAnimation.Frames[i].Duration)
		}
	}
}

func assertRealAVIFFixture(t *testing.T, source []byte, result imagecompression.Result) {
	t.Helper()

	sourceAVIF, err := avif.DecodeAll(bytes.NewReader(source), avif.Options{AutoRotate: true})
	if err != nil {
		t.Fatalf("failed to decode source AVIF: %v", err)
	}
	resultAVIF, err := avif.DecodeAll(bytes.NewReader(result.Data), avif.Options{AutoRotate: true})
	if err != nil {
		t.Fatalf("failed to decode compressed AVIF: %v", err)
	}

	if len(sourceAVIF.Image) == 0 {
		t.Fatal("expected source AVIF to contain at least one image")
	}
	if len(resultAVIF.Image) != len(sourceAVIF.Image) {
		t.Fatalf("expected %d AVIF images, got %d", len(sourceAVIF.Image), len(resultAVIF.Image))
	}
	if result.FrameCount != len(sourceAVIF.Image) {
		t.Fatalf("expected AVIF frame count %d, got %d", len(sourceAVIF.Image), result.FrameCount)
	}
	if result.Animated != (len(sourceAVIF.Image) > 1) {
		t.Fatalf("expected AVIF animated=%v, got %v", len(sourceAVIF.Image) > 1, result.Animated)
	}
}

func assertRealGIFFixture(t *testing.T, source []byte, result imagecompression.Result) {
	t.Helper()

	sourceGIF, err := gif.DecodeAll(bytes.NewReader(source))
	if err != nil {
		t.Fatalf("failed to decode source GIF animation data: %v", err)
	}
	resultGIF, err := gif.DecodeAll(bytes.NewReader(result.Data))
	if err != nil {
		t.Fatalf("failed to decode compressed GIF animation data: %v", err)
	}

	if len(resultGIF.Image) != len(sourceGIF.Image) {
		t.Fatalf("expected %d GIF frames, got %d", len(sourceGIF.Image), len(resultGIF.Image))
	}
	if result.FrameCount != len(sourceGIF.Image) {
		t.Fatalf("expected GIF frame count %d, got %d", len(sourceGIF.Image), result.FrameCount)
	}
	if result.Animated != (len(sourceGIF.Image) > 1) {
		t.Fatalf("expected GIF animated=%v, got %v", len(sourceGIF.Image) > 1, result.Animated)
	}
	if resultGIF.LoopCount != sourceGIF.LoopCount {
		t.Fatalf("expected GIF loop count %d, got %d", sourceGIF.LoopCount, resultGIF.LoopCount)
	}
	if len(resultGIF.Delay) != len(sourceGIF.Delay) {
		t.Fatalf("expected %d GIF delays, got %d", len(sourceGIF.Delay), len(resultGIF.Delay))
	}
	for i := range sourceGIF.Delay {
		if resultGIF.Delay[i] != sourceGIF.Delay[i] {
			t.Fatalf("expected GIF frame %d delay %d, got %d", i, sourceGIF.Delay[i], resultGIF.Delay[i])
		}
	}
	if len(sourceGIF.Disposal) > 0 {
		if len(resultGIF.Disposal) != len(sourceGIF.Disposal) {
			t.Fatalf("expected %d GIF disposal values, got %d", len(sourceGIF.Disposal), len(resultGIF.Disposal))
		}
		for i := range sourceGIF.Disposal {
			if resultGIF.Disposal[i] != sourceGIF.Disposal[i] {
				t.Fatalf("expected GIF frame %d disposal %d, got %d", i, sourceGIF.Disposal[i], resultGIF.Disposal[i])
			}
		}
	}
}
