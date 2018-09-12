// Package inmem provides in-memory implementations of all the domain repositories.
package inmem

import (
	"sync"

	shipping "github.com/marcusolsson/goddd"
)

type handlingEventRepository struct {
	mtx    sync.RWMutex
	events map[shipping.TrackingID][]shipping.HandlingEvent
}

// NewHandlingEventRepository returns a new instance of a in-memory handling event repository.
func NewHandlingEventRepository() shipping.HandlingEventRepository {
	return &handlingEventRepository{
		events: make(map[shipping.TrackingID][]shipping.HandlingEvent),
	}
}

func (r *handlingEventRepository) Store(e shipping.HandlingEvent) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	// Make array if it's the first event with this tracking ID.
	if _, ok := r.events[e.TrackingID]; !ok {
		r.events[e.TrackingID] = make([]shipping.HandlingEvent, 0)
	}
	r.events[e.TrackingID] = append(r.events[e.TrackingID], e)
}

func (r *handlingEventRepository) QueryHandlingHistory(id shipping.TrackingID) shipping.HandlingHistory {
	r.mtx.RLock()
	defer r.mtx.RUnlock()
	return shipping.HandlingHistory{HandlingEvents: r.events[id]}
}
