package imagehttp

import (
	"path"
	"strings"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageprocessing"
)

func cleanFilename(filename string) string {
	filename = strings.ReplaceAll(filename, "\\", "/")
	filename = path.Base(filename)

	if filename == "." || filename == "/" || filename == "" {
		return "image"
	}

	return filename
}

func Filename(filename string, format imageprocessing.Format, contentType, operation string) string {
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

	return base + "_" + operation + extension
}

func extensionMatchesFormat(extension string, format imageprocessing.Format) bool {
	switch format {
	case imageprocessing.FormatJPEG:
		return extension == ".jpg" || extension == ".jpeg"
	case imageprocessing.FormatPNG:
		return extension == ".png"
	case imageprocessing.FormatWebP:
		return extension == ".webp"
	case imageprocessing.FormatAVIF:
		return extension == ".avif"
	case imageprocessing.FormatHEIF:
		return extension == ".heic" || extension == ".heif"
	case imageprocessing.FormatGIF:
		return extension == ".gif"
	case imageprocessing.FormatBMP:
		return extension == ".bmp"
	case imageprocessing.FormatTIFF:
		return extension == ".tif" || extension == ".tiff"
	default:
		return false
	}
}

func defaultExtension(format imageprocessing.Format, contentType string) string {
	switch format {
	case imageprocessing.FormatJPEG:
		return ".jpg"
	case imageprocessing.FormatPNG:
		return ".png"
	case imageprocessing.FormatWebP:
		return ".webp"
	case imageprocessing.FormatAVIF:
		return ".avif"
	case imageprocessing.FormatHEIF:
		if contentType == "image/heif" {
			return ".heif"
		}
		return ".heic"
	case imageprocessing.FormatGIF:
		return ".gif"
	case imageprocessing.FormatBMP:
		return ".bmp"
	case imageprocessing.FormatTIFF:
		return ".tiff"
	default:
		return ".jpg"
	}
}
