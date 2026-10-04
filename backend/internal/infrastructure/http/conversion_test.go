package httpserver

import (
	"bytes"
	"golang.org/x/image/bmp"
	"image/jpeg"
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

// Codec coverage belongs to imaging tests; this small table checks transport
// naming and measured headers for each advertised destination family.
func TestConversionDestinationHeaders(t *testing.T) {
	router := NewRouter(testLogger())
	pngInput := encodePNGFixture(t, testImage(13, 7), png.BestCompression)
	var jpegInput bytes.Buffer
	if err := jpeg.Encode(&jpegInput, testImage(13, 7), nil); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct{ target, mime, extension string }{
		{"jpeg", "image/jpeg", ".jpg"}, {"png", "image/png", ".png"},
		{"webp", "image/webp", ".webp"}, {"avif", "image/avif", ".avif"},
		{"heif", "image/heic", ".heic"}, {"gif", "image/gif", ".gif"},
		{"bmp", "image/bmp", ".bmp"}, {"tiff", "image/tiff", ".tiff"},
	} {
		t.Run(tt.target, func(t *testing.T) {
			input := pngInput
			if tt.target == "png" {
				input = jpegInput.Bytes()
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, resizeRequest(t, "/images/convert", "..\\unsafe.fake", input, map[string]string{"targetFormat": tt.target}))
			if response.Code != 200 {
				t.Fatalf("%d %s", response.Code, response.Body.String())
			}
			_, params, err := mime.ParseMediaType(response.Header().Get("Content-Disposition"))
			if err != nil || params["filename"] != "unsafe_converted"+tt.extension {
				t.Fatalf("filename: %v %v", params, err)
			}
			source := "png"
			if tt.target == "png" {
				source = "jpeg"
			}
			for key, want := range map[string]string{
				"Content-Type": tt.mime, "Content-Length": strconv.Itoa(response.Body.Len()),
				"Cache-Control": "no-store", "X-Source-Format": source, "X-Output-Format": tt.target,
				"X-Image-Width": "13", "X-Image-Height": "7", "X-Original-Width": "13", "X-Original-Height": "7",
			} {
				if response.Header().Get(key) != want {
					t.Fatalf("%s: %q want %q", key, response.Header().Get(key), want)
				}
			}
		})
	}
}
