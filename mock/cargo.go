package mock

import shipping "github.com/marcusolsson/goddd"

// CargoRepository is a mock cargo repository.
type CargoRepository struct {
	StoreFn      func(c *shipping.Cargo) error
	StoreInvoked bool

	FindFn      func(id shipping.TrackingID) (*shipping.Cargo, error)
	FindInvoked bool

	FindAllFn      func() []*shipping.Cargo
	FindAllInvoked bool
}

// Store calls the StoreFn.
func (r *CargoRepository) Store(c *shipping.Cargo) error {
	r.StoreInvoked = true
	return r.StoreFn(c)
}

// Find calls the FindFn.
func (r *CargoRepository) Find(id shipping.TrackingID) (*shipping.Cargo, error) {
	r.FindInvoked = true
	return r.FindFn(id)
}

// FindAll calls the FindAllFn.
func (r *CargoRepository) FindAll() []*shipping.Cargo {
	r.FindAllInvoked = true
	return r.FindAllFn()
}
