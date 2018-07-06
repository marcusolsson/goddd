package mock

import (
	shipping "github.com/marcusolsson/goddd"
)

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

// HandlingEventRepository is a mock handling events repository.
type HandlingEventRepository struct {
	StoreFn      func(shipping.HandlingEvent)
	StoreInvoked bool

	QueryHandlingHistoryFn      func(shipping.TrackingID) shipping.HandlingHistory
	QueryHandlingHistoryInvoked bool
}

// Store calls the StoreFn.
func (r *HandlingEventRepository) Store(e shipping.HandlingEvent) {
	r.StoreInvoked = true
	r.StoreFn(e)
}

// QueryHandlingHistory calls the QueryHandlingHistoryFn.
func (r *HandlingEventRepository) QueryHandlingHistory(id shipping.TrackingID) shipping.HandlingHistory {
	r.QueryHandlingHistoryInvoked = true
	return r.QueryHandlingHistoryFn(id)
}

// RoutingService provides a mock routing service.
type RoutingService struct {
	FetchRoutesFn      func(shipping.RouteSpecification) []shipping.Itinerary
	FetchRoutesInvoked bool
}

// FetchRoutesForSpecification calls the FetchRoutesFn.
func (s *RoutingService) FetchRoutesForSpecification(rs shipping.RouteSpecification) []shipping.Itinerary {
	s.FetchRoutesInvoked = true
	return s.FetchRoutesFn(rs)
}
