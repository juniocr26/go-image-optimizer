package convert

import (
	"bytes"
	"context"
	"github.com/deepteams/webp"
	"github.com/gen2brain/avif"
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageprocessing"
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageresize"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/imaging"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/imaging/resize"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
)

type Processor struct{}

func decode(ctx context.Context, input []byte) (imageresize.Info, image.Image, error) {
	if err := ctx.Err(); err != nil {
		return imageresize.Info{}, nil, err
	}
	if len(input) == 0 {
		return imageresize.Info{}, nil, imageprocessing.ErrEmptyImage
	}
	source, err := (resize.Decoder{}).Decode(input)
	if err != nil {
		return imageresize.Info{}, nil, err
	}
	if err := ctx.Err(); err != nil {
		return imageresize.Info{}, nil, err
	}
	info := source.Info()
	img := source.(interface{ Pixels() image.Image }).Pixels()
	// JPEG, AVIF and HEIF already apply orientation while decoding.
	if img != nil && info.Format != "jpeg" && info.Format != "avif" && info.Format != "heif" {
		img = imaging.ApplyOrientation(img, imaging.ReadEXIFOrientation(input))
		info.Width, info.Height = img.Bounds().Dx(), img.Bounds().Dy()
	}
	return info, img, nil
}
func (Processor) Inspect(ctx context.Context, input []byte) (imageresize.Info, error) {
	info, _, err := decode(ctx, input)
	return info, err
}
func (Processor) Convert(ctx context.Context, input []byte, target imageprocessing.Format) (imageprocessing.Result, error) {
	if err := ctx.Err(); err != nil {
		return imageprocessing.Result{}, err
	}
	info, img, err := decode(ctx, input)
	if err != nil {
		return imageprocessing.Result{}, err
	}
	if info.FrameCount != 1 || img == nil {
		return imageprocessing.Result{}, imageprocessing.ErrUnsupportedVariant
	}
	if target == "jpeg" || target == "bmp" || target == "heif" {
		canvas := image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
		draw.Draw(canvas, canvas.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
		draw.Draw(canvas, canvas.Bounds(), img, img.Bounds().Min, draw.Over)
		img = canvas
	}
	if err := ctx.Err(); err != nil {
		return imageprocessing.Result{}, err
	}
	var out bytes.Buffer
	var data []byte
	switch target {
	case "jpeg":
		err = jpeg.Encode(&out, img, &jpeg.Options{Quality: 82})
	case "png":
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		err = encoder.Encode(&out, img)
	case "webp":
		err = webp.Encode(&out, img, &webp.EncoderOptions{Quality: 82, Method: 4, Exact: true, AlphaQuality: 100})
	case "avif":
		err = avif.Encode(&out, img, avif.Options{Quality: 60, QualityAlpha: 100, Speed: 6})
	case "heif":
		data, err = imaging.EncodeHEIF(img, 60)
	case "bmp":
		err = bmp.Encode(&out, img)
	case "tiff":
		err = tiff.Encode(&out, img, &tiff.Options{Compression: tiff.Deflate, Predictor: true})
	case "gif":
		colors := append(color.Palette{color.Transparent}, palette.WebSafe...)
		quantized := image.NewPaletted(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()), colors)
		draw.FloydSteinberg.Draw(quantized, quantized.Bounds(), img, img.Bounds().Min)
		for y := 0; y < quantized.Bounds().Dy(); y++ {
			if err := ctx.Err(); err != nil {
				return imageprocessing.Result{}, err
			}
			for x := 0; x < quantized.Bounds().Dx(); x++ {
				c := img.At(x+img.Bounds().Min.X, y+img.Bounds().Min.Y)
				_, _, _, a := c.RGBA()
				if a < 32768 {
					quantized.SetColorIndex(x, y, 0)
				} else if quantized.ColorIndexAt(x, y) == 0 {
					quantized.SetColorIndex(x, y, uint8(colors[1:].Index(c)+1))
				}
			}
		}
		err = gif.Encode(&out, quantized, nil)
	default:
		return imageprocessing.Result{}, imageprocessing.ErrUnsupportedFormat
	}
	if err != nil {
		return imageprocessing.Result{}, err
	}
	if err := ctx.Err(); err != nil {
		return imageprocessing.Result{}, err
	}
	if data == nil {
		data = out.Bytes()
	}
	if len(data) > 50<<20 {
		return imageprocessing.Result{}, imageprocessing.ErrImageTooLarge
	}
	detected, err := imaging.DetectFormat(data)
	if err != nil {
		return imageprocessing.Result{}, err
	}
	return imageprocessing.Result{SourceFormat: info.Format, Data: data, Format: detected.Format, ContentType: detected.ContentType, Width: img.Bounds().Dx(), Height: img.Bounds().Dy(), FrameCount: 1}, nil
}
