package application

import (
	"context"
)

type IUseCase[T any] interface {
	Execute(ctx context.Context, input T) error
}
