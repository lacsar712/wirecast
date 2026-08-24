package interlock

import (
	"context"
	"sync"
	"time"

	"github.com/lacsar712/wirecast/internal/model"
)

const DefaultLeaseTTL = 30 * time.Second

type lease struct {
	damper model.DamperID
	until  time.Time
}

type DamperLock struct {
	mu     sync.Mutex
	holder map[model.DamperID]lease
	clk    func() time.Time
}

func NewDamperLock(now func() time.Time) *DamperLock {
	if now == nil {
		now = time.Now
	}
	return &DamperLock{holder: make(map[model.DamperID]lease), clk: now}
}

func (l *DamperLock) TryAcquire(damper model.DamperID, ttl time.Duration) (release func(), ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.clk()
	if ex, exists := l.holder[damper]; exists && now.Before(ex.until) {
		return nil, false
	}
	until := now.Add(ttl)
	l.holder[damper] = lease{damper: damper, until: until}
	var once sync.Once
	release = func() {
		once.Do(func() {
			l.mu.Lock()
			defer l.mu.Unlock()
			if cur, ok := l.holder[damper]; ok && cur.until.Equal(until) {
				delete(l.holder, damper)
			}
		})
	}
	return release, true
}

func (l *DamperLock) WithLease(ctx context.Context, damper model.DamperID, ttl time.Duration, fn func() error) error {
	release, ok := l.TryAcquire(damper, ttl)
	if !ok {
		return model.Wrap("damper_lock", "busy", model.ErrInterlock)
	}
	defer release()
	done := make(chan error, 1)
	go func() { done <- fn() }()
	select {
	case <-ctx.Done():
		return model.Wrap("damper_lock", "canceled", context.Cause(ctx))
	case err := <-done:
		return err
	}
}

func (l *DamperLock) Held(damper model.DamperID) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	ex, ok := l.holder[damper]
	if !ok {
		return false
	}
	return l.clk().Before(ex.until)
}
