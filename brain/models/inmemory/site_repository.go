package inmemory

import "sync"

type SiteRepository struct {
	mu     sync.RWMutex
	active map[string]bool
}

func NewSiteRepository() *SiteRepository {
	return &SiteRepository{active: make(map[string]bool)}
}

func (r *SiteRepository) MarkActive(siteID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.active[siteID] = true
	return nil
}
