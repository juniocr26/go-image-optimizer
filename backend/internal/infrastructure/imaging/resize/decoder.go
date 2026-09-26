package resize

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/deepteams/webp"
	"github.com/gen2brain/avif"
	"golang.org/x/image/bmp"
	"golang.org/x/image/draw"
	"golang.org/x/image/tiff"

	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageprocessing"
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageresize"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/imaging"
)

type Decoder struct{}

type source struct {
	info   imageresize.Info
	encode func(context.Context, int, int) ([]byte, error)
}

func (s source) Info() imageresize.Info { return s.info }

func (s source) Resize(ctx context.Context, w, h int) (imageprocessing.Result, error) {
	if err := ctx.Err(); err != nil {
		return imageprocessing.Result{}, err
	}
	data, err := s.encode(ctx, w, h)
	if err != nil {
		return imageprocessing.Result{}, err
	}
	contentType := s.info.ContentType
	if detected, err := imaging.DetectFormat(data); err == nil {
		contentType = detected.ContentType
	}
	return imageprocessing.Result{Data: data, Format: s.info.Format, ContentType: contentType, Width: w, Height: h, Animated: s.info.FrameCount > 1, FrameCount: s.info.FrameCount}, nil
}

func (Decoder) Decode(input []byte) (imageresize.Source, error) {
	detected, err := imaging.DetectFormat(input)
	if err != nil {
		return nil, err
	}
	if err := validateVariant(input, detected.Format); err != nil {
		return nil, err
	}
	if detected.Format == imageprocessing.FormatGIF {
		return decodeGIF(input, detected)
	}
	if detected.Format == imageprocessing.FormatWebP {
		return decodeWebP(input, detected)
	}
	if detected.Format == imageprocessing.FormatAVIF {
		return decodeAVIF(input, detected)
	}
	var img image.Image
	if detected.Format == imageprocessing.FormatHEIF {
		img, err = imaging.DecodeHEIF(input, imageresize.MaxPixels)
	} else {
		cfg, _, configErr := image.DecodeConfig(bytes.NewReader(input))
		if configErr != nil {
			return nil, imageprocessing.ErrInvalidImage
		}
		if err := imaging.ValidateDimensions(cfg, imageresize.MaxPixels); err != nil {
			return nil, err
		}
		img, _, err = image.Decode(bytes.NewReader(input))
		if err != nil {
			return nil, imageprocessing.ErrInvalidImage
		}
		if detected.Format == imageprocessing.FormatJPEG {
			img = imaging.ApplyOrientation(img, imaging.ReadEXIFOrientation(input))
		}
	}
	if err != nil {
		return nil, err
	}
	return staticSource(img, detected, false), nil
}

func staticSource(img image.Image, detected imaging.DetectedFormat, lossless bool) source {
	return source{
		info: infoFor(img, detected, 1),
		encode: func(ctx context.Context, w, h int) ([]byte, error) {
			resized := scaleImage(img, w, h)
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			var output bytes.Buffer
			var err error
			switch detected.Format {
			case imageprocessing.FormatJPEG:
				err = jpeg.Encode(&output, resized, &jpeg.Options{Quality: 82})
			case imageprocessing.FormatPNG:
				encoder := png.Encoder{CompressionLevel: png.BestCompression}
				err = encoder.Encode(&output, resized)
			case imageprocessing.FormatBMP:
				err = bmp.Encode(&output, resized)
			case imageprocessing.FormatTIFF:
				err = tiff.Encode(&output, resized, &tiff.Options{Compression: tiff.Deflate, Predictor: true})
			case imageprocessing.FormatWebP:
				err = webp.Encode(&output, resized, &webp.EncoderOptions{Lossless: lossless, Quality: 82, Method: 4, Exact: true, AlphaQuality: 100})
			case imageprocessing.FormatHEIF:
				return imaging.EncodeHEIF(resized, 60)
			default:
				return nil, imageprocessing.ErrUnsupportedFormat
			}
			if err != nil {
				return nil, fmt.Errorf("encode resized image: %w", err)
			}
			return output.Bytes(), nil
		},
	}
}

func scaleImage(src image.Image, w, h int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	// Premultiplied alpha avoids dark fringes around transparent edges.
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Src, nil)
	return dst
}

func infoFor(img image.Image, d imaging.DetectedFormat, frames int) imageresize.Info {
	return imageresize.Info{Width: img.Bounds().Dx(), Height: img.Bounds().Dy(), Format: d.Format, ContentType: d.ContentType, FrameCount: frames}
}

func decodeAVIF(input []byte, detected imaging.DetectedFormat) (imageresize.Source, error) {
	// The installed decoder does not expose AVIF repetition counts. Reject
	// sequences instead of silently changing finite loops to infinite loops.
	if avifSequence(input) {
		return nil, imageprocessing.ErrUnsupportedVariant
	}
	cfg, err := avif.DecodeConfig(bytes.NewReader(input))
	if err != nil {
		return nil, imageprocessing.ErrInvalidImage
	}
	if err := imaging.ValidateDimensions(cfg, imageresize.MaxPixels); err != nil {
		return nil, err
	}
	anim, err := avif.DecodeAll(bytes.NewReader(input), avif.Options{AutoRotate: true})
	if err != nil || len(anim.Image) == 0 {
		return nil, imageprocessing.ErrInvalidImage
	}
	if len(anim.Image) != 1 {
		return nil, imageprocessing.ErrUnsupportedVariant
	}
	img := anim.Image[0]
	return source{info: infoFor(img, detected, 1), encode: func(ctx context.Context, w, h int) ([]byte, error) {
		resized := scaleImage(img, w, h)
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var output bytes.Buffer
		err := avif.Encode(&output, resized, avif.Options{Quality: 60, QualityAlpha: 100, Speed: 6})
		return output.Bytes(), err
	}}, nil
}
