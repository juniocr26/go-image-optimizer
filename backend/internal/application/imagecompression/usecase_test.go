package imagecompression

import (
	"context"
	"errors"
	"testing"
)

func TestUseCaseRejectsEmptyImage(t *testing.T) {
	useCase := NewUseCase(stubCompressor{})

	_, err := useCase.Execute(context.Background(), nil)

	if !errors.Is(err, ErrEmptyImage) {
		t.Fatalf("expected ErrEmptyImage, got %v", err)
	}
}

func TestUseCaseReturnsCompressedImage(t *testing.T) {
	expected := Result{
		Data:        []byte("compressed"),
		Format:      FormatPNG,
		ContentType: "image/png",
		Width:       12,
		Height:      8,
	}

	useCase := NewUseCase(stubCompressor{
		result: expected,
	})

	result, err := useCase.Execute(context.Background(), []byte("source"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if string(result.Data) != string(expected.Data) {
		t.Fatalf("expected data %q, got %q", expected.Data, result.Data)
	}

	if result.Format != expected.Format {
		t.Fatalf("expected format %q, got %q", expected.Format, result.Format)
	}

	if result.ContentType != expected.ContentType {
		t.Fatalf("expected content type %q, got %q", expected.ContentType, result.ContentType)
	}

	if result.Width != expected.Width || result.Height != expected.Height {
		t.Fatalf("expected dimensions %dx%d, got %dx%d", expected.Width, expected.Height, result.Width, result.Height)
	}
}

func TestUseCaseRespectsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	useCase := NewUseCase(stubCompressor{})

	_, err := useCase.Execute(ctx, []byte("source"))

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

type stubCompressor struct {
	result Result
	err    error
}

func (s stubCompressor) Compress([]byte) (Result, error) {
	if s.err != nil {
		return Result{}, s.err
	}

	return s.result, nil
}
