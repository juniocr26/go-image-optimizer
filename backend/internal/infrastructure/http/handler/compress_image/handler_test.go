package compressimage

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
)

func TestProcessImageRecoversFromCompressionPanic(t *testing.T) {
	request := newTestMultipartImageRequest(t, "image", "sample.png", "image/png", []byte("not empty"))
	response := httptest.NewRecorder()

	ProcessImage(testLogger(), panicUseCase{}).ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, response.Code)
	}

	if !strings.Contains(response.Body.String(), "image could not be compressed") {
		t.Fatalf("expected generic compression error, got %q", response.Body.String())
	}
}

type panicUseCase struct{}

func (panicUseCase) Execute(context.Context, []byte) (imagecompression.Result, error) {
	panic("codec panic")
}

func newTestMultipartImageRequest(t *testing.T, fieldName, filename, contentType string, payload []byte) *http.Request {
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
