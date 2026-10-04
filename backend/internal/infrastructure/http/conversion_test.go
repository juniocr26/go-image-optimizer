package httpserver

import (
	"bytes"
	"golang.org/x/image/bmp"
	"image/png"
	"mime"
	"mime/multipart"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestConversionHTTP(t *testing.T) {
	input := encodePNGFixture(t, testImage(21, 11), png.BestCompression)
	router := NewRouter(testLogger())
	response := httptest.NewRecorder()
	router.ServeHTTP(response, resizeRequest(t, "/images/convert", "../photo.jpg", input, map[string]string{"targetFormat": "bmp"}))
	if response.Code != 200 {
		t.Fatalf("%d %s", response.Code, response.Body.String())
	}
	_, params, err := mime.ParseMediaType(response.Header().Get("Content-Disposition"))
	if err != nil || params["filename"] != "photo_converted.bmp" {
		t.Fatal(params, err)
	}
	for key, want := range map[string]string{"Content-Type": "image/bmp", "Cache-Control": "no-store", "Content-Length": strconv.Itoa(response.Body.Len()), "X-Image-Width": "21", "X-Image-Height": "11", "X-Source-Format": "png", "X-Output-Format": "bmp"} {
		if response.Header().Get(key) != want {
			t.Fatal(key, response.Header())
		}
	}
	decoded, err := bmp.Decode(bytes.NewReader(response.Body.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	assertDimensions(t, decoded, 21, 11)
	if response.Body.Len() <= len(input) {
		t.Fatal("fixture should exercise larger downloadable output")
	}
	for _, options := range []map[string]string{nil, {"targetFormat": "png"}, {"targetFormat": "jpg"}, {"targetFormat": "webp", "extra": "1"}} {
		response = httptest.NewRecorder()
		router.ServeHTTP(response, resizeRequest(t, "/images/convert", "photo.png", input, options))
		if response.Code != 400 {
			t.Fatal(response.Code, response.Body.String())
		}
	}
	for _, extraFile := range []bool{false, true} {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, _ := writer.CreateFormFile("image", "photo.png")
		part.Write(input)
		writer.WriteField("targetFormat", "webp")
		if extraFile {
			part, _ = writer.CreateFormFile("image", "extra.png")
			part.Write(input)
		} else {
			writer.WriteField("targetFormat", "jpeg")
		}
		writer.Close()
		request := httptest.NewRequest("POST", "/images/convert", &body)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		response = httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != 400 {
			t.Fatal(response.Code)
		}
	}
}
