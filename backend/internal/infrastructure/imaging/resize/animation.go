package resize

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"

	"github.com/deepteams/webp"
	"github.com/deepteams/webp/animation"
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageprocessing"
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageresize"
	"github.com/juniorosa/go-image-optimizer/backend/internal/infrastructure/imaging"
)

func decodeWebP(input []byte, detected imaging.DetectedFormat) (imageresize.Source, error) {
	f, err := webp.GetFeatures(bytes.NewReader(input))
	if err != nil {
		return nil, imageprocessing.ErrInvalidImage
	}
	if err := imaging.ValidateAnimatedDimensions(f.Width, f.Height, max(1, f.FrameCount), imageresize.MaxFramePixels); err != nil {
		return nil, err
	}
	if !f.HasAnimation && f.FrameCount <= 1 {
		img, err := webp.Decode(bytes.NewReader(input))
		if err != nil {
			return nil, imageprocessing.ErrInvalidImage
		}
		return staticSource(img, detected, f.Format == "lossless"), nil
	}
	anim, err := animation.Decode(bytes.NewReader(input))
	if err != nil || len(anim.Frames) == 0 {
		return nil, imageprocessing.ErrInvalidImage
	}
	if err := imaging.ValidateAnimatedDimensions(f.Width, f.Height, len(anim.Frames), imageresize.MaxFramePixels); err != nil {
		return nil, err
	}
	if err := anim.DecodeFrames(); err != nil {
		return nil, imageprocessing.ErrInvalidImage
	}
	return source{
		info: imageresize.Info{Width: f.Width, Height: f.Height, FrameCount: len(anim.Frames), Format: detected.Format, ContentType: detected.ContentType},
		encode: func(ctx context.Context, w, h int) ([]byte, error) {
			decoder, err := animation.NewAnimDecoder(anim)
			if err != nil {
				return nil, imageprocessing.ErrInvalidImage
			}
			var output bytes.Buffer
			encoder := animation.NewEncoder(&output, w, h, &animation.EncodeOptions{LoopCount: f.LoopCount, BackgroundColor: anim.BackgroundColor, Quality: 82, Lossless: f.Format == "lossless"})
			if len(anim.ICC) > 0 {
				encoder.SetICCProfile(anim.ICC)
			}
			for decoder.HasNext() {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				frame, duration, err := decoder.NextFrame()
				if err != nil {
					return nil, imageprocessing.ErrInvalidImage
				}
				if err := encoder.AddFrame(scaleImage(frame, w, h), duration); err != nil {
					return nil, err
				}
			}
			if err := encoder.Close(); err != nil {
				return nil, err
			}
			return output.Bytes(), nil
		},
	}, nil
}

func decodeGIF(input []byte, detected imaging.DetectedFormat) (imageresize.Source, error) {
	cfg, err := gif.DecodeConfig(bytes.NewReader(input))
	if err != nil {
		return nil, imageprocessing.ErrInvalidImage
	}
	if err := imaging.ValidateDimensions(cfg, imageresize.MaxPixels); err != nil {
		return nil, err
	}
	// Count image descriptors before DecodeAll allocates frame buffers.
	frames, err := gifFrameCount(input, cfg.Width, cfg.Height)
	if err != nil {
		return nil, err
	}
	anim, err := gif.DecodeAll(bytes.NewReader(input))
	if err != nil || len(anim.Image) != frames {
		return nil, imageprocessing.ErrInvalidImage
	}
	info := imageresize.Info{Width: cfg.Width, Height: cfg.Height, FrameCount: frames, Format: detected.Format, ContentType: detected.ContentType}
	return source{info: info, encode: func(ctx context.Context, w, h int) ([]byte, error) {
		bounds := image.Rect(0, 0, cfg.Width, cfg.Height)
		canvas := image.NewRGBA(bounds)
		// A transparent source starts with a transparent canvas. Otherwise use
		// the logical screen background from the source global color table.
		background := color.Color(color.Transparent)
		transparent := false
		for _, c := range anim.Image[0].Palette {
			_, _, _, a := c.RGBA()
			transparent = transparent || a == 0
		}
		if !transparent {
			if global, ok := anim.Config.ColorModel.(color.Palette); ok && int(anim.BackgroundIndex) < len(global) {
				background = global[anim.BackgroundIndex]
			}
		}
		draw.Draw(canvas, bounds, image.NewUniform(background), image.Point{}, draw.Src)
		outputPalette := append(color.Palette{color.Transparent}, palette.WebSafe...)
		resized := &gif.GIF{LoopCount: anim.LoopCount, Delay: anim.Delay, Config: image.Config{Width: w, Height: h, ColorModel: outputPalette}, BackgroundIndex: 0}
		for i, frame := range anim.Image {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			disposal := byte(0)
			if i < len(anim.Disposal) {
				disposal = anim.Disposal[i]
			}
			var previous *image.RGBA
			if disposal == gif.DisposalPrevious {
				previous = image.NewRGBA(bounds)
				copy(previous.Pix, canvas.Pix)
			}
			draw.Draw(canvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Over)
			scaled := scaleImage(canvas, w, h)
			quantized := image.NewPaletted(scaled.Bounds(), outputPalette)
			draw.FloydSteinberg.Draw(quantized, quantized.Bounds(), scaled, image.Point{})
			// GIF has binary transparency. Quantize color separately, then
			// restore the transparent pixels so opaque dark pixels stay opaque.
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					if scaled.RGBAAt(x, y).A < 128 {
						quantized.SetColorIndex(x, y, 0)
					} else if quantized.ColorIndexAt(x, y) == 0 {
						quantized.SetColorIndex(x, y, uint8(outputPalette[1:].Index(scaled.At(x, y))+1))
					}
				}
			}
			resized.Image = append(resized.Image, quantized)
			resized.Disposal = append(resized.Disposal, gif.DisposalBackground)
			switch disposal {
			case gif.DisposalBackground:
				fill := background
				for _, c := range frame.Palette {
					_, _, _, a := c.RGBA()
					if a == 0 {
						fill = color.Transparent
						break
					}
				}
				draw.Draw(canvas, frame.Bounds(), image.NewUniform(fill), image.Point{}, draw.Src)
			case gif.DisposalPrevious:
				canvas = previous
			}
		}
		var output bytes.Buffer
		err := gif.EncodeAll(&output, resized)
		return output.Bytes(), err
	}}, nil
}

func gifFrameCount(input []byte, width, height int) (int, error) {
	if len(input) < 13 {
		return 0, imageprocessing.ErrInvalidImage
	}
	p := 13
	if input[10]&0x80 != 0 {
		p += 3 << ((input[10] & 7) + 1)
	}
	count := 0
	for p < len(input) {
		block := input[p]
		p++
		switch block {
		case 0x3b:
			if count == 0 {
				return 0, imageprocessing.ErrInvalidImage
			}
			return count, nil
		case 0x21:
			p++ // extension label; followed by data sub-blocks
		case 0x2c:
			if p+9 > len(input) {
				return 0, imageprocessing.ErrInvalidImage
			}
			packed := input[p+8]
			p += 9
			if packed&0x80 != 0 {
				p += 3 << ((packed & 7) + 1)
			}
			p++ // LZW minimum code size
			count++
			if err := imaging.ValidateAnimatedDimensions(width, height, count, imageresize.MaxFramePixels); err != nil {
				return 0, err
			}
		default:
			return 0, imageprocessing.ErrInvalidImage
		}
		for {
			if p >= len(input) {
				return 0, imageprocessing.ErrInvalidImage
			}
			size := int(input[p])
			p++
			if size == 0 {
				break
			}
			p += size
		}
	}
	return 0, imageprocessing.ErrInvalidImage
}
