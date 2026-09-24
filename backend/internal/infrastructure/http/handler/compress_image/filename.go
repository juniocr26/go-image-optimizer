package compressimage

import (
	"path"
	"strings"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imagecompression"
)

func cleanFilename(filename string) string {
	filename = strings.ReplaceAll(filename, "\\", "/")
	filename = path.Base(filename)

	if filename == "." || filename == "/" || filename == "" {
		return "image"
	}

	return filename
}

func compressedFilename(filename string, format imagecompression.Format, contentType string) string {
	filename = cleanFilename(filename)
	extension := strings.ToLower(path.Ext(filename))
	expectedExtension := defaultExtension(format, contentType)

	if !extensionMatchesFormat(extension, format) {
		extension = expectedExtension
	}

	base := strings.TrimSuffix(filename, path.Ext(filename))
	if base == "" || base == "." || base == "/" {
		base = "image"
	}

	return base + "_compressed" + extension
}

func extensionMatchesFormat(extension string, format imagecompression.Format) bool {
	switch format {
	case imagecompression.FormatJPEG:
		return extension == ".jpg" || extension == ".jpeg"
	case imagecompression.FormatPNG:
		return extension == ".png"
	case imagecompression.FormatWebP:
		return extension == ".webp"
	case imagecompression.FormatAVIF:
		return extension == ".avif"
	case imagecompression.FormatHEIF:
		return extension == ".heic" || extension == ".heif"
	case imagecompression.FormatGIF:
		return extension == ".gif"
	case imagecompression.FormatBMP:
		return extension == ".bmp"
	case imagecompression.FormatTIFF:
		return extension == ".tif" || extension == ".tiff"
	default:
		return false
	}
}

func defaultExtension(format imagecompression.Format, contentType string) string {
	switch format {
	case imagecompression.FormatJPEG:
		return ".jpg"
	case imagecompression.FormatPNG:
		return ".png"
	case imagecompression.FormatWebP:
		return ".webp"
	case imagecompression.FormatAVIF:
		return ".avif"
	case imagecompression.FormatHEIF:
		if contentType == "image/heif" {
			return ".heif"
		}
		return ".heic"
	case imagecompression.FormatGIF:
		return ".gif"
	case imagecompression.FormatBMP:
		return ".bmp"
	case imagecompression.FormatTIFF:
		return ".tiff"
	default:
		return ".jpg"
	}
}
