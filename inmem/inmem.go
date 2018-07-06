// Package inmem provides in-memory implementations of all the domain repositories.
package inmem

import (
	"sync"

	shipping "github.com/marcusolsson/goddd"
)

type cargoRepository struct {
	mtx    sync.RWMutex
	cargos map[shipping.TrackingID]*shipping.Cargo
}

func (r *cargoRepository) Store(c *shipping.Cargo) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.cargos[c.TrackingID] = c
	return nil
}

func (r *cargoRepository) Find(id shipping.TrackingID) (*shipping.Cargo, error) {
	r.mtx.RLock()
	defer r.mtx.RUnlock()
	if val, ok := r.cargos[id]; ok {
		return val, nil
	}
	return nil, shipping.ErrUnknownCargo
}

func (r *cargoRepository) FindAll() []*shipping.Cargo {
	r.mtx.RLock()
	defer r.mtx.RUnlock()
	c := make([]*shipping.Cargo, 0, len(r.cargos))
	for _, val := range r.cargos {
		c = append(c, val)
	}
	return c
}

// NewCargoRepository returns a new instance of a in-memory cargo repository.
func NewCargoRepository() shipping.CargoRepository {
	return &cargoRepository{
		cargos: make(map[shipping.TrackingID]*shipping.Cargo),
	}
}

type locationRepository struct {
	locations map[shipping.UNLocode]*shipping.Location
}

func (r *locationRepository) Find(locode shipping.UNLocode) (*shipping.Location, error) {
	if l, ok := r.locations[locode]; ok {
		return l, nil
	}
	return nil, shipping.ErrUnknownLocation
}

func (r *locationRepository) FindAll() []*shipping.Location {
	l := make([]*shipping.Location, 0, len(r.locations))
	for _, val := range r.locations {
		l = append(l, val)
	}
	return l
}

// NewLocationRepository returns a new instance of a in-memory location repository.
func NewLocationRepository() shipping.LocationRepository {
	r := &locationRepository{
		locations: make(map[shipping.UNLocode]*shipping.Location),
	}

	r.locations[shipping.SESTO] = shipping.Stockholm
	r.locations[shipping.AUMEL] = shipping.Melbourne
	r.locations[shipping.CNHKG] = shipping.Hongkong
	r.locations[shipping.JNTKO] = shipping.Tokyo
	r.locations[shipping.NLRTM] = shipping.Rotterdam
	r.locations[shipping.DEHAM] = shipping.Hamburg

	return r
}

type voyageRepository struct {
	voyages map[shipping.VoyageNumber]*shipping.Voyage
}

func (r *voyageRepository) Find(voyageNumber shipping.VoyageNumber) (*shipping.Voyage, error) {
	if v, ok := r.voyages[voyageNumber]; ok {
		return v, nil
	}

	return nil, shipping.ErrUnknownVoyage
}

// NewVoyageRepository returns a new instance of a in-memory voyage repository.
func NewVoyageRepository() shipping.VoyageRepository {
	r := &voyageRepository{
		voyages: make(map[shipping.VoyageNumber]*shipping.Voyage),
	}

	r.voyages[shipping.V100.VoyageNumber] = shipping.V100
	r.voyages[shipping.V300.VoyageNumber] = shipping.V300
	r.voyages[shipping.V400.VoyageNumber] = shipping.V400

	r.voyages[shipping.V0100S.VoyageNumber] = shipping.V0100S
	r.voyages[shipping.V0200T.VoyageNumber] = shipping.V0200T
	r.voyages[shipping.V0300A.VoyageNumber] = shipping.V0300A
	r.voyages[shipping.V0301S.VoyageNumber] = shipping.V0301S
	r.voyages[shipping.V0400S.VoyageNumber] = shipping.V0400S

	return r
}

type handlingEventRepository struct {
	mtx    sync.RWMutex
	events map[shipping.TrackingID][]shipping.HandlingEvent
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

// NewHandlingEventRepository returns a new instance of a in-memory handling event repository.
func NewHandlingEventRepository() shipping.HandlingEventRepository {
	return &handlingEventRepository{
		events: make(map[shipping.TrackingID][]shipping.HandlingEvent),
	}
}
