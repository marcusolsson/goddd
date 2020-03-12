package mock

import shipping "github.com/marcusolsson/goddd"

// VoyageRepository is a mock voyage repository.
type VoyageRepository struct {
	FindFn      func(shipping.VoyageNumber) (*shipping.Voyage, error)
	FindInvoked bool
}

// Find calls the FindFn.
func (r *VoyageRepository) Find(number shipping.VoyageNumber) (*shipping.Voyage, error) {
	r.FindInvoked = true
	return r.FindFn(number)
}
