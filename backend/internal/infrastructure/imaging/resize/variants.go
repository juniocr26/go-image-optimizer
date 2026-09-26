package resize

import (
	"bytes"
	"encoding/binary"
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageprocessing"
)

// Detect animation/container variants which the static Go decoders would
// otherwise silently flatten. Bounds checks precede every container read.
func validateVariant(input []byte, format imageprocessing.Format) error {
	if format == imageprocessing.FormatPNG {
		for p := 8; p+12 <= len(input); {
			n := uint64(binary.BigEndian.Uint32(input[p : p+4]))
			if n+12 > uint64(len(input)-p) {
				return imageprocessing.ErrInvalidImage
			}
			if bytes.Equal(input[p+4:p+8], []byte("acTL")) {
				return imageprocessing.ErrUnsupportedVariant
			}
			p += int(n) + 12
		}
	}
	if format == imageprocessing.FormatTIFF && len(input) >= 8 {
		var order binary.ByteOrder = binary.LittleEndian
		if input[0] == 'M' {
			order = binary.BigEndian
		}
		offset := uint64(order.Uint32(input[4:8]))
		if offset+2 > uint64(len(input)) {
			return imageprocessing.ErrInvalidImage
		}
		count := uint64(order.Uint16(input[offset : offset+2]))
		next := offset + 2 + 12*count
		if next+4 > uint64(len(input)) {
			return imageprocessing.ErrInvalidImage
		}
		if order.Uint32(input[next:next+4]) != 0 {
			return imageprocessing.ErrUnsupportedVariant
		}
	}
	return nil
}

func avifSequence(input []byte) bool {
	for p := 0; p+8 <= len(input); {
		n := uint64(binary.BigEndian.Uint32(input[p : p+4]))
		kind := string(input[p+4 : p+8])
		header := 8
		if n == 1 {
			if p+16 > len(input) {
				return false
			}
			n = binary.BigEndian.Uint64(input[p+8 : p+16])
			header = 16
		}
		if n == 0 {
			n = uint64(len(input) - p)
		}
		if n < uint64(header) || n > uint64(len(input)-p) {
			return false
		}
		if kind == "moov" {
			return true
		}
		if kind == "ftyp" {
			for i := p + header; i+4 <= p+int(n); i += 4 {
				if i == p+header+4 {
					continue
				} // minor version
				if string(input[i:i+4]) == "avis" {
					return true
				}
			}
		}
		p += int(n)
	}
	return false
}
