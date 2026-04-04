package txutil

import (
	"context"
	"errors"

	"github.com/agentregistry-dev/agentregistry/pkg/registry/database"
)

var ErrStoreNotConfigured = errors.New("store is not configured")

func Run(ctx context.Context, store database.Store, fn func(context.Context, database.Store) error) error {
	if store == nil {
		return ErrStoreNotConfigured
	}

	return store.InTransaction(ctx, fn)
}

func RunT[T any](ctx context.Context, store database.Store, fn func(context.Context, database.Store) (T, error)) (T, error) {
	var result T
	var fnErr error

	err := Run(ctx, store, func(txCtx context.Context, txStore database.Store) error {
		result, fnErr = fn(txCtx, txStore)
		return fnErr
	})
	if err != nil {
		var zero T
		return zero, err
	}

	return result, nil
}
