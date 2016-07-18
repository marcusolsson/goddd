package mock

import (
	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/voyage"
)

// CargoRepository is a mock cargo repository.
type CargoRepository struct {
	StoreFn   func(c *cargo.Cargo) error
	FindFn    func(id cargo.TrackingID) (*cargo.Cargo, error)
	FindAllFn func() []*cargo.Cargo
}

// Store calls the StoreFn.
func (r *CargoRepository) Store(c *cargo.Cargo) error {
	return r.StoreFn(c)
}

// Find calls the FindFn.
func (r *CargoRepository) Find(id cargo.TrackingID) (*cargo.Cargo, error) {
	return r.FindFn(id)
}

// FindAll calls the FindAllFn.
func (r *CargoRepository) FindAll() []*cargo.Cargo {
	return r.FindAllFn()
}

// LocationRepository is a mock location repository.
type LocationRepository struct {
	FindFn    func(location.UNLocode) (location.Location, error)
	FindAllFn func() []location.Location
}

// Find calls the FindFn.
func (r *LocationRepository) Find(locode location.UNLocode) (location.Location, error) {
	return r.FindFn(locode)
}

// FindAll calls the FindAllFn.
func (r *LocationRepository) FindAll() []location.Location {
	return r.FindAllFn()
}

// VoyageRepository is a mock voyage repository.
type VoyageRepository struct {
	FindFn func(voyage.Number) (*voyage.Voyage, error)
}

// Find calls the FindFn.
func (r *VoyageRepository) Find(number voyage.Number) (*voyage.Voyage, error) {
	return r.FindFn(number)
}

// HandlingEventRepository is a mock handling events repository.
type HandlingEventRepository struct {
	StoreFn                func(cargo.HandlingEvent)
	QueryHandlingHistoryFn func(cargo.TrackingID) cargo.HandlingHistory
}

// Store calls the StoreFn.
func (r *HandlingEventRepository) Store(e cargo.HandlingEvent) {
	r.StoreFn(e)
}

// QueryHandlingHistory calls the QueryHandlingHistoryFn.
func (r *HandlingEventRepository) QueryHandlingHistory(id cargo.TrackingID) cargo.HandlingHistory {
	return r.QueryHandlingHistoryFn(id)
}

// RoutingService provides a mock routing service.
type RoutingService struct {
	FetchRoutesFn func(cargo.RouteSpecification) []cargo.Itinerary
}

// FetchRoutesForSpecification calls the FetchRoutesFn.
func (s *RoutingService) FetchRoutesForSpecification(rs cargo.RouteSpecification) []cargo.Itinerary {
	return s.FetchRoutesFn(rs)
}
