package inmem

import (
	"sync"

	shipping "github.com/marcusolsson/goddd"
)

type cargoRepository struct {
	mtx    sync.RWMutex
	cargos map[shipping.TrackingID]*shipping.Cargo
}

// NewCargoRepository returns a new instance of a in-memory cargo repository.
func NewCargoRepository() shipping.CargoRepository {
	return &cargoRepository{
		cargos: make(map[shipping.TrackingID]*shipping.Cargo),
	}
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
