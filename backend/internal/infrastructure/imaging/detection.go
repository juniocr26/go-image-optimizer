package imaging

import (
	"bytes"
	"encoding/binary"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageprocessing"
)

type DetectedFormat struct {
	Format      imageprocessing.Format
	ContentType string
}

func DetectFormat(input []byte) (DetectedFormat, error) {
	if isUnsupportedCameraRAW(input) {
		return DetectedFormat{}, imageprocessing.ErrUnsupportedFormat
	}

	switch {
	case hasPrefix(input, []byte{0xff, 0xd8, 0xff}):
		return DetectedFormat{Format: imageprocessing.FormatJPEG, ContentType: "image/jpeg"}, nil
	case hasPrefix(input, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}):
		return DetectedFormat{Format: imageprocessing.FormatPNG, ContentType: "image/png"}, nil
	case hasPrefix(input, []byte("GIF87a")) || hasPrefix(input, []byte("GIF89a")):
		return DetectedFormat{Format: imageprocessing.FormatGIF, ContentType: "image/gif"}, nil
	case hasPrefix(input, []byte("BM")):
		return DetectedFormat{Format: imageprocessing.FormatBMP, ContentType: "image/bmp"}, nil
	case hasPrefix(input, []byte{'I', 'I', '*', 0}) || hasPrefix(input, []byte{'M', 'M', 0, '*'}):
		return DetectedFormat{Format: imageprocessing.FormatTIFF, ContentType: "image/tiff"}, nil
	case isWebP(input):
		return DetectedFormat{Format: imageprocessing.FormatWebP, ContentType: "image/webp"}, nil
	}

	if format, ok := detectBMFFImageFormat(input); ok {
		return format, nil
	}

	return DetectedFormat{}, imageprocessing.ErrUnsupportedFormat
}

func hasPrefix(input, prefix []byte) bool {
	return len(input) >= len(prefix) && bytes.Equal(input[:len(prefix)], prefix)
}

func isWebP(input []byte) bool {
	return len(input) >= 12 &&
		bytes.Equal(input[:4], []byte("RIFF")) &&
		bytes.Equal(input[8:12], []byte("WEBP"))
}

func isUnsupportedCameraRAW(input []byte) bool {
	return isCanonCR2(input) ||
		isCanonCR3(input) ||
		isDigitalNegative(input) ||
		hasPrefix(input, []byte("FUJIFILMCCD-RAW")) ||
		hasPrefix(input, []byte{'I', 'I', 'U', 0})
}

func isCanonCR2(input []byte) bool {
	return len(input) >= 12 &&
		hasTIFFHeader(input) &&
		bytes.Equal(input[8:12], []byte{'C', 'R', 0x02, 0x00})
}

func isCanonCR3(input []byte) bool {
	brands, ok := bmffBrands(input)
	return ok && hasBrand(brands, "crx ")
}

func isDigitalNegative(input []byte) bool {
	byteOrder, firstIFDOffset, ok := tiffByteOrder(input)
	if !ok {
		return false
	}

	offset := int(firstIFDOffset)
	if uint32(offset) != firstIFDOffset || offset+2 > len(input) {
		return false
	}

	entryCount := int(byteOrder.Uint16(input[offset : offset+2]))
	entryOffset := offset + 2

	for i := 0; i < entryCount; i++ {
		if entryOffset+12 > len(input) {
			return false
		}

		tag := byteOrder.Uint16(input[entryOffset : entryOffset+2])
		if tag == 0xc612 {
			return true
		}

		entryOffset += 12
	}

	return false
}

func tiffByteOrder(input []byte) (binary.ByteOrder, uint32, bool) {
	if len(input) < 8 {
		return nil, 0, false
	}

	switch {
	case bytes.Equal(input[:4], []byte{'I', 'I', '*', 0}):
		return binary.LittleEndian, binary.LittleEndian.Uint32(input[4:8]), true
	case bytes.Equal(input[:4], []byte{'M', 'M', 0, '*'}):
		return binary.BigEndian, binary.BigEndian.Uint32(input[4:8]), true
	default:
		return nil, 0, false
	}
}

func hasTIFFHeader(input []byte) bool {
	_, _, ok := tiffByteOrder(input)
	return ok
}

func detectBMFFImageFormat(input []byte) (DetectedFormat, bool) {
	brands, ok := bmffBrands(input)
	if !ok {
		return DetectedFormat{}, false
	}

	if hasBrand(brands, "avif", "avis") {
		return DetectedFormat{Format: imageprocessing.FormatAVIF, ContentType: "image/avif"}, true
	}

	if hasBrand(brands, "heic", "heix", "hevc", "hevx", "heim", "heis", "hevm", "hevs") {
		return DetectedFormat{Format: imageprocessing.FormatHEIF, ContentType: "image/heic"}, true
	}

	if hasBrand(brands, "heif", "mif1", "msf1") {
		return DetectedFormat{Format: imageprocessing.FormatHEIF, ContentType: "image/heif"}, true
	}

	return DetectedFormat{}, false
}

func bmffBrands(input []byte) (map[string]struct{}, bool) {
	if len(input) < 12 || !bytes.Equal(input[4:8], []byte("ftyp")) {
		return nil, false
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

	return brands, true
}

func hasBrand(brands map[string]struct{}, candidates ...string) bool {
	for _, candidate := range candidates {
		if _, ok := brands[candidate]; ok {
			return true
		}
	}

	return false
}
