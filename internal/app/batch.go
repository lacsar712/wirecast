package app

import (
	"context"

	"github.com/lacsar712/wirecast/internal/model"
)

func (a *App) BeginBatchScope(ctx context.Context, tower model.TowerID) (context.Context, context.CancelFunc) {
	if tower == "" {
		tower = model.TowerID(a.cfg.TowerID)
	}
	a.batchMu.Lock()
	if cancel, ok := a.batchCancels[tower]; ok {
		cancel()
	}
	child, cancel := context.WithCancel(ctx)
	a.batchCancels[tower] = cancel
	a.batchMu.Unlock()
	release := func() {
		a.batchMu.Lock()
		delete(a.batchCancels, tower)
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
