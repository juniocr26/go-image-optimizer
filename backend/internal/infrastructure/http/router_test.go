package httpserver

import (
	"bytes"
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

func TestProcessImageReturnsUploadedBytes(t *testing.T) {
	imageBytes := []byte("\x89PNG\r\n\x1a\nsample image bytes")
	request := newMultipartImageRequest(t, "image", "sample.png", "image/png", imageBytes)
	response := httptest.NewRecorder()

	NewRouter(testLogger()).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %q", http.StatusOK, response.Code, response.Body.String())
	}

	if !bytes.Equal(response.Body.Bytes(), imageBytes) {
		t.Fatalf("response body does not match uploaded image bytes")
	}

	if got := response.Header().Get("Content-Type"); got != "image/png" {
		t.Fatalf("expected Content-Type image/png, got %q", got)
	}

	if got := response.Header().Get("Content-Length"); got != strconv.Itoa(len(imageBytes)) {
		t.Fatalf("expected Content-Length %d, got %q", len(imageBytes), got)
	}

	if got := response.Header().Get("Content-Disposition"); !strings.Contains(got, "sample.png") {
		t.Fatalf("expected Content-Disposition to include filename, got %q", got)
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
