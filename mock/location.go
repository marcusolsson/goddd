package mock

import shipping "github.com/marcusolsson/goddd"

// LocationRepository is a mock location repository.
type LocationRepository struct {
	FindFn      func(shipping.UNLocode) (*shipping.Location, error)
	FindInvoked bool

	FindAllFn      func() []*shipping.Location
	FindAllInvoked bool
}

// Find calls the FindFn.
func (r *LocationRepository) Find(locode shipping.UNLocode) (*shipping.Location, error) {
	r.FindInvoked = true
	return r.FindFn(locode)
}

// FindAll calls the FindAllFn.
func (r *LocationRepository) FindAll() []*shipping.Location {
	r.FindAllInvoked = true
	return r.FindAllFn()
}
