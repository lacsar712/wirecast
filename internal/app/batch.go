package app

import (
	"context"

	"github.com/lacsar712/wirecast/internal/model"
)

func (a *App) BeginBatchScope(ctx context.Context, tower model.TowerID) (context.Context, context.CancelFunc) {
	child, cancel := context.WithCancel(ctx)
	a.batchMu.Lock()
	if a.activeCancel != nil {
		a.activeCancel()
	}
	a.activeCancel = cancel
	a.batchMu.Unlock()
	release := func() {
		a.batchMu.Lock()
		a.activeCancel = nil
		a.batchMu.Unlock()
		cancel()
	}
	return child, release
}

func (a *App) RunBatch(ctx context.Context, tower model.TowerID, fn func(context.Context) error) error {
	batchCtx, release := a.BeginBatchScope(ctx, tower)
	defer release()
	return fn(batchCtx)
}
