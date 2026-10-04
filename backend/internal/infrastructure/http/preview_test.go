package httpserver

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"mime/multipart"
	"net/http/httptest"
	"testing"
)

func TestPreviewHTTP(t *testing.T) {
	input := encodePNGFixture(t, testImage(21, 11), png.BestCompression)
	before := bytes.Clone(input)
	router := NewRouter(testLogger())
	response := httptest.NewRecorder()
	router.ServeHTTP(response, resizeRequest(t, "/images/preview", "spoof.heic", input, nil))
	if response.Code != 200 {
		t.Fatal(response.Code, response.Body.String())
	}
	decoded, format, err := image.Decode(bytes.NewReader(response.Body.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	assertDimensions(t, decoded, 21, 11)
	if response.Header().Get("Content-Type") != "image/"+format || response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("Content-Disposition") != "" {
		t.Fatal(response.Header())
	}
	if !bytes.Equal(input, before) {
		t.Fatal("changed original")
	}
	for _, tc := range []struct {
		input   []byte
		options map[string]string
		status  int
	}{{[]byte("bad"), nil, 415}, {nil, nil, 400}, {input, map[string]string{"extra": "1"}, 400}} {
		response = httptest.NewRecorder()
		router.ServeHTTP(response, resizeRequest(t, "/images/preview", "photo", tc.input, tc.options))
		if response.Code != tc.status {
			t.Fatal(response.Code, response.Body.String())
		}
		var body map[string]string
		if json.Unmarshal(response.Body.Bytes(), &body) != nil || body["error"] == "" {
			t.Fatal(response.Body.String())
		}
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for i := 0; i < 2; i++ {
		part, _ := writer.CreateFormFile("image", "photo")
		part.Write(input)
	}
	writer.Close()
	request := httptest.NewRequest("POST", "/images/preview", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != 400 {
		t.Fatal(response.Code)
	}
}
