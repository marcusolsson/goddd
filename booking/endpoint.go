package booking

import (
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/routing"
	"golang.org/x/net/context"
)

type bookCargoRequest struct {
	Origin          location.UNLocode
	Destination     location.UNLocode
	ArrivalDeadline time.Time
}

type bookCargoResponse struct {
	ID cargo.TrackingID `json:"tracking_id"`
}

func makeBookCargoEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(bookCargoRequest)

		id, err := s.BookNewCargo(req.Origin, req.Destination, req.ArrivalDeadline)
		if err != nil {
			return nil, err
		}

		return bookCargoResponse{ID: id}, nil
	}
}

type requestRoutesRequest struct {
	ID cargo.TrackingID
}

type requestRoutesResponse struct {
	Routes []routing.Route `json:"routes"`
}

func makeRequestRoutesEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(requestRoutesRequest)

		itineraries := s.RequestPossibleRoutesForCargo(req.ID)

		result := []routing.Route{}
		for _, itin := range itineraries {
			var legs []routing.Leg
			for _, leg := range itin.Legs {
				legs = append(legs, routing.Leg{
					VoyageNumber: string(leg.VoyageNumber),
					From:         string(leg.LoadLocation),
					To:           string(leg.UnloadLocation),
					LoadTime:     leg.LoadTime,
					UnloadTime:   leg.UnloadTime,
				})
			}
			result = append(result, routing.Route{Legs: legs})
		}

		return requestRoutesResponse{Routes: result}, nil
	}
}

type assignToRouteRequest struct {
	ID        cargo.TrackingID
	Itinerary cargo.Itinerary
}

type assignToRouteResponse struct {
}

func makeAssignToRouteEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(assignToRouteRequest)

		if err := s.AssignCargoToRoute(req.ID, req.Itinerary); err != nil {
			return nil, err
		}

		return assignToRouteResponse{}, nil
	}
}

type changeDestinationRequest struct {
	ID          cargo.TrackingID
	Destination location.UNLocode
}

type changeDestinationResponse struct {
}

func makeChangeDestinationEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(changeDestinationRequest)

		if err := s.ChangeDestination(req.ID, req.Destination); err != nil {
			return nil, err
		}

		return changeDestinationResponse{}, nil
	}
}

type listCargosRequest struct{}

type listCargosResponse struct {
	Cargos []Cargo `json:"cargos"`
}

func makeListCargosEndpoint(bs Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return listCargosResponse{Cargos: bs.Cargos()}, nil
	}
}

type listLocationsRequest struct {
}

type listLocationsResponse struct {
	Locations []Location `json:"locations"`
}

func makeListLocationsEndpoint(bs Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		_ = request.(listLocationsRequest)
		return listLocationsResponse{Locations: bs.Locations()}, nil
	}
}

type fetchRoutesRequest struct {
	From string
	To   string
}

type fetchRoutesResponse struct {
	Paths []struct {
		Edges []struct {
			Origin      string    `json:"origin"`
			Destination string    `json:"destination"`
			Voyage      string    `json:"voyage"`
			Departure   time.Time `json:"departure"`
			Arrival     time.Time `json:"arrival"`
		} `json:"edges"`
	} `json:"paths"`
}
