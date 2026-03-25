package leriantest

import (
	"github.com/LerianStudio/lerian-sdk-golang/models"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

func fakeInjectedError(cfg *fakeConfig, key string) error {
	if cfg == nil || key == "" {
		return nil
	}
	return cfg.injectedError(key)
}

func fakeGetStored[T any](cfg *fakeConfig, injectedKey, operation, resource, id string, store *fakeStore[T]) (*T, error) {
	if err := fakeInjectedError(cfg, injectedKey); err != nil {
		return nil, err
	}

	item, ok := store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound(operation, resource, id)
	}

	return &item, nil
}

func fakeListStored[T any](cfg *fakeConfig, injectedKey string, store *fakeStore[T], opts *models.CursorListOptions) *pagination.Iterator[T] {
	if err := fakeInjectedError(cfg, injectedKey); err != nil {
		return pagination.NewErrorIterator[T](err)
	}

	return store.PaginatedIterator(opts)
}

func fakeDeleteStored[T any](cfg *fakeConfig, injectedKey, operation, resource, id string, store *fakeStore[T]) error {
	if err := fakeInjectedError(cfg, injectedKey); err != nil {
		return err
	}

	if _, ok := store.Get(id); !ok {
		return sdkerrors.NewNotFound(operation, resource, id)
	}

	store.Delete(id)
	return nil
}

func fakeActionStored[T any](cfg *fakeConfig, injectedKey, operation, resource, id string, store *fakeStore[T]) (*T, error) {
	if err := fakeInjectedError(cfg, injectedKey); err != nil {
		return nil, err
	}

	item, ok := store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound(operation, resource, id)
	}

	return &item, nil
}

func fakeMutateStored[T any](cfg *fakeConfig, injectedKey, operation, resource, id string, store *fakeStore[T], mutate func(*T)) (*T, error) {
	item, err := fakeActionStored(cfg, injectedKey, operation, resource, id, store)
	if err != nil {
		return nil, err
	}

	mutate(item)
	store.Set(id, *item)
	return item, nil
}

func fakeScopedMutateStored[T any](cfg *fakeConfig, injectedKey, operation, resource, id string, store *fakeStore[T], inScope func(T) bool, mutate func(*T) error) (*T, error) {
	if err := fakeInjectedError(cfg, injectedKey); err != nil {
		return nil, err
	}

	item, ok := store.Get(id)
	if !ok || !inScope(item) {
		return nil, sdkerrors.NewNotFound(operation, resource, id)
	}

	if err := mutate(&item); err != nil {
		return nil, err
	}

	store.Set(id, item)
	return &item, nil
}

func fakeStaticResult[T any](cfg *fakeConfig, injectedKey string, result T) (*T, error) {
	if err := fakeInjectedError(cfg, injectedKey); err != nil {
		return nil, err
	}

	return &result, nil
}
