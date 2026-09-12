package imaging

import (
	"bytes"
	"encoding/binary"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
)

type detectedFormat struct {
	format      imagecompression.Format
	contentType string
}

func detectFormat(input []byte) (detectedFormat, error) {
	switch {
	case hasPrefix(input, []byte{0xff, 0xd8, 0xff}):
		return detectedFormat{format: imagecompression.FormatJPEG, contentType: "image/jpeg"}, nil
	case hasPrefix(input, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}):
		return detectedFormat{format: imagecompression.FormatPNG, contentType: "image/png"}, nil
	case hasPrefix(input, []byte("GIF87a")) || hasPrefix(input, []byte("GIF89a")):
		return detectedFormat{format: imagecompression.FormatGIF, contentType: "image/gif"}, nil
	case hasPrefix(input, []byte("BM")):
		return detectedFormat{format: imagecompression.FormatBMP, contentType: "image/bmp"}, nil
	case hasPrefix(input, []byte{'I', 'I', '*', 0}) || hasPrefix(input, []byte{'M', 'M', 0, '*'}):
		return detectedFormat{format: imagecompression.FormatTIFF, contentType: "image/tiff"}, nil
	case isWebP(input):
		return detectedFormat{format: imagecompression.FormatWebP, contentType: "image/webp"}, nil
	}

	if format, ok := detectBMFFImageFormat(input); ok {
		return format, nil
	}

	return detectedFormat{}, imagecompression.ErrUnsupportedFormat
}

func hasPrefix(input, prefix []byte) bool {
	return len(input) >= len(prefix) && bytes.Equal(input[:len(prefix)], prefix)
}

func isWebP(input []byte) bool {
	return len(input) >= 12 &&
		bytes.Equal(input[:4], []byte("RIFF")) &&
		bytes.Equal(input[8:12], []byte("WEBP"))
}

func detectBMFFImageFormat(input []byte) (detectedFormat, bool) {
	if len(input) < 12 || !bytes.Equal(input[4:8], []byte("ftyp")) {
		return detectedFormat{}, false
	}

	boxSize := int(binary.BigEndian.Uint32(input[:4]))
	end := len(input)
	if boxSize >= 16 && boxSize <= len(input) {
		end = boxSize
	} else if end > 128 {
		end = 128
	}

	brands := make(map[string]struct{})
	for offset := 8; offset+4 <= end; offset += 4 {
		brands[string(input[offset:offset+4])] = struct{}{}
	}

	if hasBrand(brands, "avif", "avis") {
		return detectedFormat{format: imagecompression.FormatAVIF, contentType: "image/avif"}, true
	}

	if hasBrand(brands, "heic", "heix", "hevc", "hevx", "heim", "heis", "hevm", "hevs") {
		return detectedFormat{format: imagecompression.FormatHEIF, contentType: "image/heic"}, true
	}

	if hasBrand(brands, "heif", "mif1", "msf1") {
		return detectedFormat{format: imagecompression.FormatHEIF, contentType: "image/heif"}, true
	}

	return detectedFormat{}, false
}

func hasBrand(brands map[string]struct{}, candidates ...string) bool {
	for _, candidate := range candidates {
		if _, ok := brands[candidate]; ok {
			return true
		}
	}

	return false
}
