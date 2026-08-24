package interlock

import (
	"sync"
	"time"

	"github.com/lacsar712/wirecast/internal/model"
)

type feedLease struct {
	holder string
	until  time.Time
}

type FeedLeaseRegistry struct {
	mu   sync.Mutex
	now  func() time.Time
	held map[model.TowerID]feedLease
}

func NewFeedLeaseRegistry(now func() time.Time) *FeedLeaseRegistry {
	if now == nil {
		now = time.Now
	}
	return &FeedLeaseRegistry{now: now, held: make(map[model.TowerID]feedLease)}
}

func (r *FeedLeaseRegistry) Require(tower model.TowerID, holder string, ttl time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := r.now()
	if ex, ok := r.held[tower]; ok && now.Before(ex.until) && ex.holder != holder {
		return model.Wrap("feed_lease", "busy", model.ErrInterlock)
	}
	r.held[tower] = feedLease{holder: holder, until: now.Add(ttl)}
	return nil
}

func (r *FeedLeaseRegistry) ReleaseHolder(tower model.TowerID, holder string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if ex, ok := r.held[tower]; ok && ex.holder == holder {
		delete(r.held, tower)
	}
}

func (r *FeedLeaseRegistry) HeldByOther(tower model.TowerID, holder string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	ex, ok := r.held[tower]
	if !ok {
		return false
	}
	return r.now().Before(ex.until) && ex.holder != holder
}
