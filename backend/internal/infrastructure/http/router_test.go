package httpserver

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strconv"
	"strings"
	"testing"
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

func TestProcessImageCompressesJPEG(t *testing.T) {
	imageBytes := encodeJPEGFixture(t, 80, 60)
	request := newMultipartImageRequest(t, "image", "sample.jpg", "image/jpeg", imageBytes)
	response := httptest.NewRecorder()

	NewRouter(testLogger()).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %q", http.StatusOK, response.Code, response.Body.String())
	}

	decoded, format, err := image.Decode(bytes.NewReader(response.Body.Bytes()))
	if err != nil {
		t.Fatalf("response body is not a decodable image: %v", err)
	}

	if format != "jpeg" {
		t.Fatalf("expected decoded format jpeg, got %q", format)
	}

	if decoded.Bounds().Dx() != 80 || decoded.Bounds().Dy() != 60 {
		t.Fatalf("expected dimensions 80x60, got %dx%d", decoded.Bounds().Dx(), decoded.Bounds().Dy())
	}

	if got := response.Header().Get("Content-Type"); got != "image/jpeg" {
		t.Fatalf("expected Content-Type image/jpeg, got %q", got)
	}

	if got := response.Header().Get("Content-Length"); got != strconv.Itoa(response.Body.Len()) {
		t.Fatalf("expected Content-Length %d, got %q", response.Body.Len(), got)
	}

	if got := response.Header().Get("Content-Disposition"); !strings.Contains(got, "sample_compressed.jpg") {
		t.Fatalf("expected Content-Disposition to include compressed filename, got %q", got)
	}
}

func TestProcessImageCompressesPNG(t *testing.T) {
	imageBytes := encodePNGFixture(t, 48, 32)
	request := newMultipartImageRequest(t, "image", "sample.png", "image/png", imageBytes)
	response := httptest.NewRecorder()

	NewRouter(testLogger()).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %q", http.StatusOK, response.Code, response.Body.String())
	}

	decoded, format, err := image.Decode(bytes.NewReader(response.Body.Bytes()))
	if err != nil {
		t.Fatalf("response body is not a decodable image: %v", err)
	}

	if format != "png" {
		t.Fatalf("expected decoded format png, got %q", format)
	}

	if decoded.Bounds().Dx() != 48 || decoded.Bounds().Dy() != 32 {
		t.Fatalf("expected dimensions 48x32, got %dx%d", decoded.Bounds().Dx(), decoded.Bounds().Dy())
	}

	if got := response.Header().Get("Content-Type"); got != "image/png" {
		t.Fatalf("expected Content-Type image/png, got %q", got)
	}

	if got := response.Header().Get("Content-Disposition"); !strings.Contains(got, "sample_compressed.png") {
		t.Fatalf("expected Content-Disposition to include compressed filename, got %q", got)
	}
}

func TestProcessImageNormalizesMismatchedFilenameExtension(t *testing.T) {
	imageBytes := encodeJPEGFixture(t, 24, 24)
	request := newMultipartImageRequest(t, "image", "sample.png", "image/png", imageBytes)
	response := httptest.NewRecorder()

	NewRouter(testLogger()).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %q", http.StatusOK, response.Code, response.Body.String())
	}

	if got := response.Header().Get("Content-Disposition"); !strings.Contains(got, "sample_compressed.jpg") {
		t.Fatalf("expected Content-Disposition to use actual image format, got %q", got)
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

	if !strings.Contains(response.Body.String(), "unsupported image format") {
		t.Fatalf("expected unsupported format error, got %q", response.Body.String())
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
	oversizedPayload := bytes.Repeat([]byte("a"), 25<<20)
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

func encodeJPEGFixture(t *testing.T, width, height int) []byte {
	t.Helper()

	var buffer bytes.Buffer
	if err := jpeg.Encode(&buffer, testImage(width, height), &jpeg.Options{Quality: 100}); err != nil {
		t.Fatalf("failed to encode jpeg fixture: %v", err)
	}

	return buffer.Bytes()
}

func encodePNGFixture(t *testing.T, width, height int) []byte {
	t.Helper()

	var buffer bytes.Buffer
	encoder := png.Encoder{CompressionLevel: png.NoCompression}
	if err := encoder.Encode(&buffer, testImage(width, height)); err != nil {
		t.Fatalf("failed to encode png fixture: %v", err)
	}

	return buffer.Bytes()
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
