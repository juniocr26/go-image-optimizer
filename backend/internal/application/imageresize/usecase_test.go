package imageresize

import (
	"context"
	"errors"
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageprocessing"
	"testing"
)

func TestTargetDimensions(t *testing.T) {
	info := Info{Width: 899, Height: 1599, FrameCount: 1}
	tests := []struct {
		name    string
		options Options
		w, h    int
		err     error
	}{
		{"half with odd dimensions", Options{Mode: "percentage", Reduction: 50}, 450, 800, nil},
		{"quarter smaller", Options{Mode: "percentage", Reduction: 25}, 674, 1199, nil},
		{"three quarters smaller", Options{Mode: "percentage", Reduction: 75}, 225, 400, nil},
		{"width anchors ratio", Options{Mode: "pixels", Width: 450, Height: 999, Axis: "width", KeepAspectRatio: true}, 450, 800, nil},
		{"height anchors ratio", Options{Mode: "pixels", Width: 999, Height: 800, Axis: "height", KeepAspectRatio: true}, 450, 800, nil},
		{"stretch", Options{Mode: "pixels", Width: 100, Height: 100, Axis: "width"}, 100, 100, nil},
		{"enlarge", Options{Mode: "pixels", Width: 1798, Height: 3198, Axis: "width", KeepAspectRatio: true}, 1798, 3198, nil},
		{"cap proportional enlargement", Options{Mode: "pixels", Width: 1798, Height: 3198, Axis: "width", KeepAspectRatio: true, WithoutEnlargement: true}, 899, 1599, nil},
		{"cap independent dimensions", Options{Mode: "pixels", Width: 1798, Height: 200, Axis: "width", WithoutEnlargement: true}, 899, 200, nil},
		{"invalid mode", Options{Mode: "crop"}, 0, 0, ErrInvalidOptions},
		{"invalid percentage", Options{Mode: "percentage", Reduction: 0}, 0, 0, ErrInvalidOptions},
		{"negative dimension", Options{Mode: "pixels", Width: -1, Height: 10, Axis: "width"}, 0, 0, ErrInvalidOptions},
		{"zero dimension", Options{Mode: "pixels", Width: 1, Height: 0, Axis: "width"}, 0, 0, ErrInvalidOptions},
		{"missing axis", Options{Mode: "pixels", Width: 1, Height: 1}, 0, 0, ErrInvalidOptions},
		{"unsafe dimensions", Options{Mode: "pixels", Width: 10000, Height: 10000, Axis: "width"}, 0, 0, imageprocessing.ErrImageTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, h, err := Target(info, tt.options)
			if w != tt.w || h != tt.h || !errors.Is(err, tt.err) {
				t.Fatalf("got %dx%d / %v, want %dx%d / %v", w, h, err, tt.w, tt.h, tt.err)
			}
		})
	}
	if w, h, err := Target(Info{Width: 1, Height: 1, FrameCount: 1}, Options{Mode: "percentage", Reduction: 75}); err != nil || w != 1 || h != 1 {
		t.Fatalf("minimum pixel: %d %d %v", w, h, err)
	}
	_, _, err := Target(Info{Width: 10, Height: 10, FrameCount: 100}, Options{Mode: "pixels", Width: 1000, Height: 1000, Axis: "width"})
	if !errors.Is(err, imageprocessing.ErrAnimationTooLarge) {
		t.Fatalf("expected frame pixel rejection, got %v", err)
	}
	_, _, err = Target(Info{Width: 1, Height: 32_000_000, FrameCount: 1}, Options{Mode: "pixels", Width: 32_000_000, Height: 1, Axis: "width", KeepAspectRatio: true})
	if !errors.Is(err, imageprocessing.ErrImageTooLarge) {
		t.Fatalf("hostile ratio: %v", err)
	}
}

type stubDecoder struct {
	called bool
	source Source
	err    error
}

func (s *stubDecoder) Decode([]byte) (Source, error) { s.called = true; return s.source, s.err }

type stubSource struct{ called bool }

func (*stubSource) Info() Info {
	return Info{Width: 20, Height: 10, FrameCount: 1, Format: imageprocessing.FormatPNG, ContentType: "image/png"}
}
func (s *stubSource) Resize(context.Context, int, int) (imageprocessing.Result, error) {
	s.called = true
	return imageprocessing.Result{Width: 10, Height: 5, Data: []byte("resized")}, nil
}

func TestUseCaseNoOpAndCancellation(t *testing.T) {
	src := &stubSource{}
	d := &stubDecoder{source: src}
	uc := NewUseCase(d)
	input := []byte("original")
	result, err := uc.Execute(context.Background(), input, Options{Mode: "pixels", Width: 20, Height: 10, Axis: "width"})
	if err != nil || src.called || string(result.Data) != "original" || result.OriginalWidth != 20 {
		t.Fatalf("no-op: %+v %v", result, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	d.called = false
	_, err = uc.Inspect(ctx, input)
	if !errors.Is(err, context.Canceled) || d.called {
		t.Fatalf("canceled inspect: %v", err)
	}
	_, err = uc.Execute(ctx, input, Options{Mode: "percentage", Reduction: 50})
	if !errors.Is(err, context.Canceled) || d.called {
		t.Fatalf("canceled execute: %v", err)
	}
	_, err = uc.Inspect(context.Background(), nil)
	if !errors.Is(err, imageprocessing.ErrEmptyImage) {
		t.Fatalf("empty: %v", err)
	}
	_, err = uc.Execute(context.Background(), input, Options{Mode: "percentage", Reduction: 999})
	if !errors.Is(err, ErrInvalidOptions) || d.called {
		t.Fatalf("invalid options should precede decoding: %v", err)
	}
	d.err = imageprocessing.ErrInvalidImage
	_, err = uc.Execute(context.Background(), input, Options{Mode: "percentage", Reduction: 50})
	if !errors.Is(err, d.err) {
		t.Fatalf("decode error: %v", err)
	}
}
