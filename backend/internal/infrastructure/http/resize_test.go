package httpserver

import (
	"bytes"
	"encoding/json"
	"image/png"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func resizeRequest(t *testing.T, path, filename string, input []byte, options map[string]string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if filename != "" {
		part, err := writer.CreateFormFile("image", filename)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(input); err != nil {
			t.Fatal(err)
		}
	}
	for key, value := range options {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, path, &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}

func TestResizeHTTPMetadataAndDownload(t *testing.T) {
	input := encodePNGFixture(t, testImage(21, 11), png.NoCompression)
	router := NewRouter(testLogger())
	response := httptest.NewRecorder()
	router.ServeHTTP(response, resizeRequest(t, "/images/resize/info", "photo.jpg", input, nil))
	if response.Code != 200 {
		t.Fatalf("inspect: %d %s", response.Code, response.Body.String())
	}
	var info struct {
		Width, Height, FrameCount int
		ContentType               string
	}
	if err := json.Unmarshal(response.Body.Bytes(), &info); err != nil {
		t.Fatal(err)
	}
	if info.Width != 21 || info.Height != 11 || info.FrameCount != 1 || info.ContentType != "image/png" {
		t.Fatalf("metadata: %+v", info)
	}
	response = httptest.NewRecorder()
	router.ServeHTTP(response, resizeRequest(t, "/images/resize", "../photo.jpg", input, map[string]string{"mode": "percentage", "reduction": "50"}))
	if response.Code != 200 {
		t.Fatalf("resize: %d %s", response.Code, response.Body.String())
	}
	if response.Header().Get("Content-Type") != "image/png" || response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("invalid response headers")
	}
	_, params, err := mime.ParseMediaType(response.Header().Get("Content-Disposition"))
	if err != nil {
		t.Fatal(err)
	}
	if params["filename"] != "photo_resized.png" {
		t.Fatalf("unsafe/spoofed filename: %s", params["filename"])
	}
	for key, want := range map[string]string{"X-Image-Width": "11", "X-Image-Height": "6", "X-Original-Width": "21", "X-Original-Height": "11", "Content-Length": strconv.Itoa(response.Body.Len())} {
		if response.Header().Get(key) != want {
			t.Fatalf("%s: %q", key, response.Header().Get(key))
		}
	}
	decoded := decodePNG(t, response.Body.Bytes())
	assertDimensions(t, decoded, 11, 6)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, resizeRequest(t, "/images/resize", "photo.png", input, map[string]string{"mode": "pixels", "width": "42", "height": "22"}))
	if response.Code != 200 || !bytes.Equal(response.Body.Bytes(), input) {
		t.Fatal("default no-enlarge should return original bytes")
	}
}

func TestResizeHTTPValidation(t *testing.T) {
	input := encodePNGFixture(t, testImage(20, 10), png.NoCompression)
	cases := []struct {
		name    string
		file    string
		data    []byte
		options map[string]string
		status  int
	}{
		{"missing image", "", nil, map[string]string{"mode": "percentage", "reduction": "50"}, 400},
		{"missing options", "image.png", input, nil, 400},
		{"invalid mode", "image.png", input, map[string]string{"mode": "crop"}, 400},
		{"fractional pixels", "image.png", input, map[string]string{"mode": "pixels", "width": "4.5", "height": "4"}, 400},
		{"negative pixels", "image.png", input, map[string]string{"mode": "pixels", "width": "-1", "height": "4"}, 400},
		{"overflow", "image.png", input, map[string]string{"mode": "pixels", "width": "999999999999999999999999999999", "height": "4"}, 400},
		{"NaN", "image.png", input, map[string]string{"mode": "pixels", "width": "NaN", "height": "4"}, 400},
		{"empty bool", "image.png", input, map[string]string{"mode": "percentage", "reduction": "50", "withoutEnlargement": ""}, 400},
		{"invalid bool", "image.png", input, map[string]string{"mode": "percentage", "reduction": "50", "keepAspectRatio": "maybe"}, 400},
		{"invalid reduction", "image.png", input, map[string]string{"mode": "percentage", "reduction": "150"}, 400},
		{"invalid axis", "image.png", input, map[string]string{"mode": "pixels", "width": "4", "height": "4", "axis": "depth"}, 400},
		{"output budget", "image.png", input, map[string]string{"mode": "pixels", "width": "10000", "height": "10000", "keepAspectRatio": "false", "withoutEnlargement": "false"}, 413},
		{"multi-page TIFF", "image.tiff", []byte{'I', 'I', 42, 0, 8, 0, 0, 0, 0, 0, 14, 0, 0, 0, 0, 0, 0, 0}, map[string]string{"mode": "percentage", "reduction": "50"}, 422},
		{"unknown format", "image.png", []byte("hello"), map[string]string{"mode": "percentage", "reduction": "50"}, 415},
		{"corrupt PNG", "image.png", []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}, map[string]string{"mode": "percentage", "reduction": "50"}, 400},
		{"empty", "image.png", nil, map[string]string{"mode": "percentage", "reduction": "50"}, 400},
		{"upload budget", "image.png", make([]byte, 50<<20), map[string]string{"mode": "percentage", "reduction": "50"}, 413},
	}
	router := NewRouter(testLogger())
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, resizeRequest(t, "/images/resize", tt.file, tt.data, tt.options))
			if response.Code != tt.status {
				t.Fatalf("got %d want %d: %s", response.Code, tt.status, response.Body.String())
			}
		})
	}
	for _, path := range []string{"/images/resize", "/images/resize/info"} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString("bad"))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(response, request)
		if response.Code != 400 {
			t.Fatal("accepted non-multipart")
		}
		response = httptest.NewRecorder()
		request = httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString("bad"))
		request.Header.Set("Content-Type", "multipart/form-data; boundary=test")
		router.ServeHTTP(response, request)
		if response.Code != 400 {
			t.Fatal("accepted malformed multipart")
		}
	}
}

func TestResizeHTTPRejectsAmbiguousMultipart(t *testing.T) {
	input := encodePNGFixture(t, testImage(20, 10), png.NoCompression)
	for _, path := range []string{"/images/resize", "/images/resize/info"} {
		for _, extra := range []string{"duplicate image", "extra file", "text image", "duplicate option"} {
			if path == "/images/resize/info" && extra == "duplicate option" {
				continue
			}
			t.Run(path+"/"+extra, func(t *testing.T) {
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)
				part, err := writer.CreateFormFile("image", "photo.png")
				if err != nil {
					t.Fatal(err)
				}
				if _, err = part.Write(input); err != nil {
					t.Fatal(err)
				}
				if err = writer.WriteField("mode", "percentage"); err != nil {
					t.Fatal(err)
				}
				if err = writer.WriteField("reduction", "50"); err != nil {
					t.Fatal(err)
				}
				switch extra {
				case "duplicate image", "extra file":
					field := "image"
					if extra == "extra file" {
						field = "other"
					}
					part, err = writer.CreateFormFile(field, "second.png")
					if err != nil {
						t.Fatal(err)
					}
					if _, err = part.Write(input); err != nil {
						t.Fatal(err)
					}
				case "text image":
					if err = writer.WriteField("image", "ambiguous"); err != nil {
						t.Fatal(err)
					}
				case "duplicate option":
					if err = writer.WriteField("reduction", "25"); err != nil {
						t.Fatal(err)
					}
				}
				if err = writer.Close(); err != nil {
					t.Fatal(err)
				}
				req := httptest.NewRequest(http.MethodPost, path, &body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				response := httptest.NewRecorder()
				NewRouter(testLogger()).ServeHTTP(response, req)
				if response.Code != http.StatusBadRequest {
					t.Fatalf("got %d: %s", response.Code, response.Body.String())
				}
			})
		}
	}
}
