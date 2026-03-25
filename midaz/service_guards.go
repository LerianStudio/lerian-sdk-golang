package midaz

import (
	"context"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

func ensureService[T any](svc *T) error {
	if svc == nil {
		return core.ErrNilService
	}

	return nil
}

func newErrorIterator[T any](err error) *pagination.Iterator[T] {
	return pagination.NewIterator[T](func(_ context.Context, _ string) ([]T, string, error) {
		return nil, "", err
	})
}
