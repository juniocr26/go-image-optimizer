package imageconversion

import (
	"context"
	"errors"
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageprocessing"
	"github.com/juniorosa/go-image-optimizer/backend/internal/application/imageresize"
)

var ErrInvalidTarget = errors.New("destination must be a supported format different from the source")
var Targets = []imageprocessing.Format{"jpeg", "png", "webp", "avif", "heif", "gif", "bmp", "tiff"}

func ValidTarget(target imageprocessing.Format) bool {
	for _, f := range Targets {
		if f == target {
			return true
		}
	}
	return false
}

type Processor interface {
	Inspect(context.Context, []byte) (imageresize.Info, error)
	Convert(context.Context, []byte, imageprocessing.Format) (imageprocessing.Result, error)
}
type UseCase struct{ processor Processor }

func NewUseCase(p Processor) UseCase { return UseCase{p} }
func (u UseCase) Inspect(ctx context.Context, input []byte) (imageresize.Info, error) {
	return u.processor.Inspect(ctx, input)
}
func (u UseCase) Execute(ctx context.Context, input []byte, target imageprocessing.Format) (imageprocessing.Result, error) {
	if !ValidTarget(target) {
		return imageprocessing.Result{}, ErrInvalidTarget
	}
	info, err := u.Inspect(ctx, input)
	if err != nil {
		return imageprocessing.Result{}, err
	}
	if info.Format == target {
		return imageprocessing.Result{}, ErrInvalidTarget
	}
	if info.FrameCount != 1 {
		return imageprocessing.Result{}, imageprocessing.ErrUnsupportedVariant
	}
	return u.processor.Convert(ctx, input, target)
}
