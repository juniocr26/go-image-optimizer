package httpserver

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/deepteams/webp"
	"github.com/gen2brain/avif"
	"github.com/strukturag/libheif/go/heif"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

func TestHealth(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	NewRouter(testLogger()).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	if strings.TrimSpace(response.Body.String()) != `{"status":"ok"}` {
		t.Fatalf("unexpected response body: %q", response.Body.String())
	}
}

func TestProcessImageCompressesSupportedFormats(t *testing.T) {
	source := testImage(40, 28)

	tests := []struct {
		name                string
		filename            string
		requestContentType  string
		input               []byte
		responseContentType string
		downloadName        string
		decode              func(*testing.T, []byte) image.Image
	}{
		{
			name:                "jpeg",
			filename:            "sample.jpg",
			requestContentType:  "image/jpeg",
			input:               encodeJPEGFixture(t, source, 100),
			responseContentType: "image/jpeg",
			downloadName:        "sample_compressed.jpg",
			decode:              decodeJPEG,
		},
		{
			name:                "png",
			filename:            "sample.png",
			requestContentType:  "image/png",
			input:               encodePNGFixture(t, source, png.NoCompression),
			responseContentType: "image/png",
			downloadName:        "sample_compressed.png",
			decode:              decodePNG,
		},
		{
			name:                "webp",
			filename:            "sample.webp",
			requestContentType:  "image/webp",
			input:               encodeWebPFixture(t, source),
			responseContentType: "image/webp",
			downloadName:        "sample_compressed.webp",
			decode:              decodeWebP,
		},
		{
			name:                "avif",
			filename:            "sample.avif",
			requestContentType:  "image/avif",
			input:               encodeAVIFFixture(t, source),
			responseContentType: "image/avif",
			downloadName:        "sample_compressed.avif",
			decode:              decodeAVIF,
		},
		{
			name:                "heic",
			filename:            "sample.heic",
			requestContentType:  "image/heic",
			input:               encodeHEIFFixture(t, source),
			responseContentType: "image/heic",
			downloadName:        "sample_compressed.heic",
			decode:              decodeHEIF,
		},
		{
			name:                "heif-extension",
			filename:            "sample.heif",
			requestContentType:  "image/heif",
			input:               encodeHEIFFixture(t, source),
			responseContentType: "image/heic",
			downloadName:        "sample_compressed.heif",
			decode:              decodeHEIF,
		},
		{
			name:                "gif",
			filename:            "sample.gif",
			requestContentType:  "image/gif",
			input:               encodeGIFFixture(t, source),
			responseContentType: "image/gif",
			downloadName:        "sample_compressed.gif",
			decode:              decodeGIF,
		},
		{
			name:                "bmp",
			filename:            "sample.bmp",
			requestContentType:  "image/bmp",
			input:               encodeBMPFixture(t, source),
			responseContentType: "image/bmp",
			downloadName:        "sample_compressed.bmp",
			decode:              decodeBMP,
		},
		{
			name:                "tiff",
			filename:            "sample.tiff",
			requestContentType:  "image/tiff",
			input:               encodeTIFFFixture(t, source),
			responseContentType: "image/tiff",
			downloadName:        "sample_compressed.tiff",
			decode:              decodeTIFF,
		},
		{
			name:                "tif-extension",
			filename:            "sample.tif",
			requestContentType:  "image/tiff",
			input:               encodeTIFFFixture(t, source),
			responseContentType: "image/tiff",
			downloadName:        "sample_compressed.tif",
			decode:              decodeTIFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := newMultipartImageRequest(t, "image", tt.filename, tt.requestContentType, tt.input)
			response := httptest.NewRecorder()

			NewRouter(testLogger()).ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d with body %q", http.StatusOK, response.Code, response.Body.String())
			}

			decoded := tt.decode(t, response.Body.Bytes())
			assertDimensions(t, decoded, 40, 28)

			if got := response.Header().Get("Content-Type"); got != tt.responseContentType {
				t.Fatalf("expected Content-Type %q, got %q", tt.responseContentType, got)
			}

			if got := response.Header().Get("Content-Length"); got != strconv.Itoa(response.Body.Len()) {
				t.Fatalf("expected Content-Length %d, got %q", response.Body.Len(), got)
			}

			if got := response.Header().Get("Content-Disposition"); !strings.Contains(got, tt.downloadName) {
				t.Fatalf("expected Content-Disposition to include %q, got %q", tt.downloadName, got)
			}
		})
	}
}

func TestProcessImageNormalizesMismatchedFilenameExtensionsFromContent(t *testing.T) {
	source := testImage(24, 24)

	tests := []struct {
		name         string
		filename     string
		contentType  string
		input        []byte
		downloadName string
	}{
		{
			name:         "jpeg named png",
			filename:     "sample.png",
			contentType:  "image/png",
			input:        encodeJPEGFixture(t, source, 100),
			downloadName: "sample_compressed.jpg",
		},
		{
			name:         "png named jpg",
			filename:     "sample.jpg",
			contentType:  "image/jpeg",
			input:        encodePNGFixture(t, source, png.NoCompression),
			downloadName: "sample_compressed.png",
		},
		{
			name:         "webp named bmp",
			filename:     "sample.bmp",
			contentType:  "image/bmp",
			input:        encodeWebPFixture(t, source),
			downloadName: "sample_compressed.webp",
		},
		{
			name:         "avif named jpg",
			filename:     "sample.jpg",
			contentType:  "image/jpeg",
			input:        encodeAVIFFixture(t, source),
			downloadName: "sample_compressed.avif",
		},
		{
			name:         "heic named png",
			filename:     "sample.png",
			contentType:  "image/png",
			input:        encodeHEIFFixture(t, source),
			downloadName: "sample_compressed.heic",
		},
		{
			name:         "gif named webp",
			filename:     "sample.webp",
			contentType:  "image/webp",
			input:        encodeGIFFixture(t, source),
			downloadName: "sample_compressed.gif",
		},
		{
			name:         "bmp named tiff",
			filename:     "sample.tiff",
			contentType:  "image/tiff",
			input:        encodeBMPFixture(t, source),
			downloadName: "sample_compressed.bmp",
		},
		{
			name:         "tiff named jpeg",
			filename:     "sample.jpeg",
			contentType:  "image/jpeg",
			input:        encodeTIFFFixture(t, source),
			downloadName: "sample_compressed.tiff",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := newMultipartImageRequest(t, "image", tt.filename, tt.contentType, tt.input)
			response := httptest.NewRecorder()

			NewRouter(testLogger()).ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d with body %q", http.StatusOK, response.Code, response.Body.String())
			}

			if got := response.Header().Get("Content-Disposition"); !strings.Contains(got, tt.downloadName) {
				t.Fatalf("expected Content-Disposition to include %q, got %q", tt.downloadName, got)
			}
		})
	}
}

func TestProcessImageRequiresImageField(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	field, err := writer.CreateFormField("not_image")
	if err != nil {
		t.Fatalf("failed to create form field: %v", err)
	}

	if _, err := field.Write([]byte("value")); err != nil {
		t.Fatalf("failed to write form field: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/images/compress", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response := httptest.NewRecorder()

	NewRouter(testLogger()).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}

	if !strings.Contains(response.Body.String(), "image field is required") {
		t.Fatalf("expected missing image error, got %q", response.Body.String())
	}
}

func TestProcessImageRejectsMalformedMultipart(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/images/compress", strings.NewReader("not a valid multipart body"))
	request.Header.Set("Content-Type", "multipart/form-data; boundary=broken")
	response := httptest.NewRecorder()

	NewRouter(testLogger()).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}

	if !strings.Contains(response.Body.String(), "invalid multipart form") {
		t.Fatalf("expected malformed multipart error, got %q", response.Body.String())
	}
}

func TestProcessImageRequiresMultipartForm(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/images/compress", strings.NewReader("{}"))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	NewRouter(testLogger()).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}

	if !strings.Contains(response.Body.String(), "request must be multipart/form-data") {
		t.Fatalf("expected multipart error, got %q", response.Body.String())
	}
}

func TestProcessImageRejectsUnsupportedFormat(t *testing.T) {
	request := newMultipartImageRequest(t, "image", "notes.txt", "text/plain", []byte("not an image"))
	response := httptest.NewRecorder()

	NewRouter(testLogger()).ServeHTTP(response, request)

	if response.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected status %d, got %d", http.StatusUnsupportedMediaType, response.Code)
	}

	if !strings.Contains(response.Body.String(), "Use JPEG, PNG, WebP, AVIF, HEIC, GIF, BMP, or TIFF") {
		t.Fatalf("expected supported-format error, got %q", response.Body.String())
	}
}

func TestProcessImageRejectsUnsupportedRAWInput(t *testing.T) {
	request := newMultipartImageRequest(t, "image", "photo.tiff", "image/tiff", minimalDNGFixture())
	response := httptest.NewRecorder()

	NewRouter(testLogger()).ServeHTTP(response, request)

	if response.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected status %d, got %d", http.StatusUnsupportedMediaType, response.Code)
	}

	if !strings.Contains(response.Body.String(), "unsupported image format") {
		t.Fatalf("expected unsupported format error, got %q", response.Body.String())
	}
}

func TestProcessImageDoesNotTrustFilenameExtension(t *testing.T) {
	request := newMultipartImageRequest(t, "image", "broken.webp", "image/webp", []byte("not a webp"))
	response := httptest.NewRecorder()

	NewRouter(testLogger()).ServeHTTP(response, request)

	if response.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected status %d, got %d", http.StatusUnsupportedMediaType, response.Code)
	}

	if !strings.Contains(response.Body.String(), "unsupported image format") {
		t.Fatalf("expected unsupported format error, got %q", response.Body.String())
	}
}

func TestProcessImageRejectsEmptyImage(t *testing.T) {
	request := newMultipartImageRequest(t, "image", "empty.png", "image/png", nil)
	response := httptest.NewRecorder()

	NewRouter(testLogger()).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}

	if !strings.Contains(response.Body.String(), "uploaded image is empty") {
		t.Fatalf("expected empty image error, got %q", response.Body.String())
	}
}

func TestProcessImageRejectsCorruptedImage(t *testing.T) {
	corruptedPNG := []byte("\x89PNG\r\n\x1a\nnot a valid png body")
	request := newMultipartImageRequest(t, "image", "broken.png", "image/png", corruptedPNG)
	response := httptest.NewRecorder()

	NewRouter(testLogger()).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}

	if !strings.Contains(response.Body.String(), "image content is invalid or corrupted") {
		t.Fatalf("expected corrupted image error, got %q", response.Body.String())
	}
}

func TestProcessImageRejectsRequestsAboveUploadLimit(t *testing.T) {
	oversizedPayload := bytes.Repeat([]byte("a"), 50<<20)
	request := newMultipartImageRequest(t, "image", "large.png", "image/png", oversizedPayload)
	response := httptest.NewRecorder()

	NewRouter(testLogger()).ServeHTTP(response, request)

	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status %d, got %d", http.StatusRequestEntityTooLarge, response.Code)
	}

	if !strings.Contains(response.Body.String(), "uploaded image is too large") {
		t.Fatalf("expected size limit error, got %q", response.Body.String())
	}
}

func TestProcessImageAcceptsValidImageNearUploadLimit(t *testing.T) {
	payload := largePNGFixture(t, (50<<20)-4096)
	request := newMultipartImageRequest(t, "image", "near-limit.png", "image/png", payload)
	response := httptest.NewRecorder()

	NewRouter(testLogger()).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %q", http.StatusOK, response.Code, response.Body.String())
	}

	if got := response.Header().Get("Content-Type"); got != "image/png" {
		t.Fatalf("expected Content-Type %q, got %q", "image/png", got)
	}

	decoded := decodePNG(t, response.Body.Bytes())
	assertDimensions(t, decoded, 1, 1)
}

func newMultipartImageRequest(t *testing.T, fieldName, filename, contentType string, payload []byte) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	headers := make(textproto.MIMEHeader)
	headers.Set("Content-Disposition", `form-data; name="`+fieldName+`"; filename="`+filename+`"`)
	headers.Set("Content-Type", contentType)

	part, err := writer.CreatePart(headers)
	if err != nil {
		t.Fatalf("failed to create image part: %v", err)
	}

	if _, err := part.Write(payload); err != nil {
		t.Fatalf("failed to write image part: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/images/compress", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	return request
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func testImage(width, height int) image.Image {
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

func encodeJPEGFixture(t *testing.T, img image.Image, quality int) []byte {
	t.Helper()

	var buffer bytes.Buffer
	if err := jpeg.Encode(&buffer, img, &jpeg.Options{Quality: quality}); err != nil {
		t.Fatalf("failed to encode jpeg fixture: %v", err)
	}

	return buffer.Bytes()
}

func encodePNGFixture(t *testing.T, img image.Image, level png.CompressionLevel) []byte {
	t.Helper()

	var buffer bytes.Buffer
	encoder := png.Encoder{CompressionLevel: level}
	if err := encoder.Encode(&buffer, img); err != nil {
		t.Fatalf("failed to encode png fixture: %v", err)
	}

	return buffer.Bytes()
}

func encodeWebPFixture(t *testing.T, img image.Image) []byte {
	t.Helper()

	var buffer bytes.Buffer
	if err := webp.Encode(&buffer, img, &webp.EncoderOptions{Quality: 100, Method: 4}); err != nil {
		t.Fatalf("failed to encode webp fixture: %v", err)
	}

	return buffer.Bytes()
}

func encodeAVIFFixture(t *testing.T, img image.Image) []byte {
	t.Helper()

	var buffer bytes.Buffer
	if err := avif.Encode(&buffer, img, avif.Options{Quality: 70, QualityAlpha: 100, Speed: 8}); err != nil {
		t.Fatalf("failed to encode avif fixture: %v", err)
	}

	return buffer.Bytes()
}

func encodeHEIFFixture(t *testing.T, img image.Image) []byte {
	t.Helper()

	ctx, err := heif.EncodeFromImage(toRGBAFixture(img), heif.CompressionHEVC, 70, heif.LosslessModeDisabled, heif.LoggingLevelNone)
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

func encodeGIFFixture(t *testing.T, img image.Image) []byte {
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

func encodeBMPFixture(t *testing.T, img image.Image) []byte {
	t.Helper()

	var buffer bytes.Buffer
	if err := bmp.Encode(&buffer, img); err != nil {
		t.Fatalf("failed to encode bmp fixture: %v", err)
	}

	return buffer.Bytes()
}

func encodeTIFFFixture(t *testing.T, img image.Image) []byte {
	t.Helper()

	var buffer bytes.Buffer
	if err := tiff.Encode(&buffer, img, &tiff.Options{Compression: tiff.Uncompressed}); err != nil {
		t.Fatalf("failed to encode tiff fixture: %v", err)
	}

	return buffer.Bytes()
}

func largePNGFixture(t *testing.T, targetSize int) []byte {
	t.Helper()

	base := encodePNGFixture(t, testImage(1, 1), png.BestCompression)
	iendTypeOffset := bytes.LastIndex(base, []byte("IEND"))
	if iendTypeOffset < 4 {
		t.Fatal("png fixture does not contain IEND chunk")
	}

	chunkStart := iendTypeOffset - 4
	textPrefix := []byte("Comment\x00")
	fillerSize := targetSize - len(base) - 12 - len(textPrefix)
	if fillerSize < 0 {
		t.Fatalf("target PNG size %d is too small", targetSize)
	}

	text := make([]byte, 0, len(textPrefix)+fillerSize)
	text = append(text, textPrefix...)
	text = append(text, bytes.Repeat([]byte("x"), fillerSize)...)
	textChunk := pngChunk("tEXt", text)

	output := make([]byte, 0, len(base)+len(textChunk))
	output = append(output, base[:chunkStart]...)
	output = append(output, textChunk...)
	output = append(output, base[chunkStart:]...)

	if len(output) != targetSize {
		t.Fatalf("expected large PNG size %d, got %d", targetSize, len(output))
	}

	return output
}

func pngChunk(chunkType string, data []byte) []byte {
	chunk := make([]byte, 12+len(data))
	binary.BigEndian.PutUint32(chunk[:4], uint32(len(data)))
	copy(chunk[4:8], chunkType)
	copy(chunk[8:8+len(data)], data)
	binary.BigEndian.PutUint32(chunk[8+len(data):], crc32.ChecksumIEEE(chunk[4:8+len(data)]))

	return chunk
}

func minimalDNGFixture() []byte {
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

func decodeJPEG(t *testing.T, data []byte) image.Image {
	t.Helper()

	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to decode jpeg response: %v", err)
	}

	return img
}

func decodePNG(t *testing.T, data []byte) image.Image {
	t.Helper()

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to decode png response: %v", err)
	}

	return img
}

func decodeWebP(t *testing.T, data []byte) image.Image {
	t.Helper()

	img, err := webp.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to decode webp response: %v", err)
	}

	return img
}

func decodeAVIF(t *testing.T, data []byte) image.Image {
	t.Helper()

	img, err := avif.Decode(bytes.NewReader(data), avif.Options{AutoRotate: true})
	if err != nil {
		t.Fatalf("failed to decode avif response: %v", err)
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
		t.Fatalf("failed to read heif response: %v", err)
	}
	handle, err := ctx.GetPrimaryImageHandle()
	if err != nil {
		t.Fatalf("failed to get heif primary handle: %v", err)
	}
	decoded, err := handle.DecodeImage(heif.ColorspaceRGB, heif.ChromaInterleavedRGBA, nil)
	if err != nil {
		t.Fatalf("failed to decode heif response: %v", err)
	}
	img, err := decoded.GetImage()
	if err != nil {
		t.Fatalf("failed to convert heif response: %v", err)
	}

	return img
}

func decodeGIF(t *testing.T, data []byte) image.Image {
	t.Helper()

	img, err := gif.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to decode gif response: %v", err)
	}

	return img
}

func decodeBMP(t *testing.T, data []byte) image.Image {
	t.Helper()

	img, err := bmp.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to decode bmp response: %v", err)
	}

	return img
}

func decodeTIFF(t *testing.T, data []byte) image.Image {
	t.Helper()

	img, err := tiff.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to decode tiff response: %v", err)
	}

	return img
}

func toRGBAFixture(img image.Image) *image.RGBA {
	bounds := img.Bounds()
	rgba := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x-bounds.Min.X, y-bounds.Min.Y, img.At(x, y))
		}
	}

	return rgba
}

func assertDimensions(t *testing.T, img image.Image, width, height int) {
	t.Helper()

	bounds := img.Bounds()
	if bounds.Dx() != width || bounds.Dy() != height {
		t.Fatalf("expected dimensions %dx%d, got %dx%d", width, height, bounds.Dx(), bounds.Dy())
	}
}
