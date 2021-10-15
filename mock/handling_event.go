package mock

import (
	shipping "github.com/marcusolsson/goddd"
)

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
