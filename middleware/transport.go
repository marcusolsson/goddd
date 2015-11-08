package middleware

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/gorilla/mux"
	"github.com/marcusolsson/goddd/booking"
	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/handling"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/routing"
	"github.com/marcusolsson/goddd/tracking"
	"github.com/marcusolsson/goddd/voyage"
)

// EncodeResponse encodes a response to JSON.
func EncodeResponse(w http.ResponseWriter, resp interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(resp)
}

// List cargos

type listCargosRequest struct{}

type listCargosResponse struct {
	Cargos []booking.Cargo `json:"cargos"`
}

// MakeListCargosEndpoint returns a endpoint to list cargos.
func MakeListCargosEndpoint(bs booking.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return listCargosResponse{Cargos: bs.Cargos()}, nil
	}
}

// DecodeListCargosRequest returns a decoded request to list cargos.
func DecodeListCargosRequest(r *http.Request) (interface{}, error) {
	return listCargosRequest{}, nil
}

// Find cargo

type findCargoRequest struct {
	ID string
}

type findCargoResponse struct {
	Cargo tracking.Cargo `json:"cargo"`
}

// MakeFindCargoEndpoint returns a endpoint to find a cargo.
func MakeFindCargoEndpoint(ts tracking.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(findCargoRequest)

		c, err := ts.Track(req.ID)
		if err != nil {
			return nil, err
		}

		return findCargoResponse{Cargo: c}, nil
	}
}

// DecodeFindCargoRequest returns a decoded request to find a cargo.
func DecodeFindCargoRequest(r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)

	id := vars["id"]
	if id == "" {
		return nil, errors.New("missing parameters")
	}

	return findCargoRequest{ID: id}, nil
}

// Book cargo

type bookCargoRequest struct {
	Origin          location.UNLocode
	Destination     location.UNLocode
	ArrivalDeadline time.Time
}

type bookCargoResponse struct {
	ID cargo.TrackingID `json:"trackingId"`
}

// MakeBookCargoEndpoint returns a endpoint to book a cargo.
func MakeBookCargoEndpoint(s booking.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(bookCargoRequest)

		id, err := s.BookNewCargo(req.Origin, req.Destination, req.ArrivalDeadline)
		if err != nil {
			return nil, err
		}

		return bookCargoResponse{ID: id}, nil
	}
}

// DecodeBookCargoRequest returns a decoded request to find a cargo.
func DecodeBookCargoRequest(r *http.Request) (interface{}, error) {
	var (
		origin          = r.URL.Query().Get("origin")
		destination     = r.URL.Query().Get("destination")
		arrivalDeadline = r.URL.Query().Get("arrivalDeadline")
	)

	if origin == "" || destination == "" || arrivalDeadline == "" {
		return nil, errors.New("missing parameters")
	}

	millis, _ := strconv.ParseInt(arrivalDeadline, 10, 64)

	return bookCargoRequest{
		Origin:          location.UNLocode(origin),
		Destination:     location.UNLocode(destination),
		ArrivalDeadline: time.Unix(millis/1000, 0),
	}, nil
}

// Request routes

type requestRoutesRequest struct {
	ID cargo.TrackingID
}

type requestRoutesResponse struct {
	Routes []routing.Route `json:"routes"`
}

// MakeRequestRoutesEndpoint returns a endpoint to request routes.
func MakeRequestRoutesEndpoint(s booking.Service) endpoint.Endpoint {
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

// DecodeRequestRoutesRequest returns a decoded request to request routes.
func DecodeRequestRoutesRequest(r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)

	id := vars["id"]
	if id == "" {
		return nil, errors.New("missing parameters")
	}

	return requestRoutesRequest{ID: cargo.TrackingID(id)}, nil
}

// Assign to route

type assignToRouteRequest struct {
	ID        cargo.TrackingID
	Itinerary cargo.Itinerary
}

type assignToRouteResponse struct {
}

// MakeAssignToRouteEndpoint returns a endpoint to assign a cargo to a route.
func MakeAssignToRouteEndpoint(s booking.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(assignToRouteRequest)

		if err := s.AssignCargoToRoute(req.ID, req.Itinerary); err != nil {
			return nil, err
		}

		return assignToRouteResponse{}, nil
	}
}

// DecodeAssignToRouteRequest returns a decoded request to assign cargo to a
// route.
func DecodeAssignToRouteRequest(r *http.Request) (interface{}, error) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var route routing.Route
	if err := json.Unmarshal(b, &route); err != nil {
		return nil, err
	}

	var legs []cargo.Leg
	for _, l := range route.Legs {
		legs = append(legs, cargo.Leg{
			VoyageNumber:   voyage.Number(l.VoyageNumber),
			LoadLocation:   location.UNLocode(l.From),
			UnloadLocation: location.UNLocode(l.To),
			LoadTime:       l.LoadTime,
			UnloadTime:     l.UnloadTime,
		})
	}

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		return nil, errors.New("missing parameters")
	}

	return assignToRouteRequest{ID: cargo.TrackingID(id), Itinerary: cargo.Itinerary{Legs: legs}}, nil
}

// Change destination

type changeDestinationRequest struct {
	ID          cargo.TrackingID
	Destination location.UNLocode
}

type changeDestinationResponse struct {
}

// DecodeChangeDestinationRequest returns a decoded request to change
// destination of a cargo.
func DecodeChangeDestinationRequest(r *http.Request) (interface{}, error) {
	id := mux.Vars(r)["id"]
	destination := r.URL.Query().Get("destination")

	if id == "" || destination == "" {
		return nil, errors.New("missing parameters")
	}

	return changeDestinationRequest{
		ID:          cargo.TrackingID(id),
		Destination: location.UNLocode(destination),
	}, nil
}

// MakeChangeDestinationEndpoint returns a endpoint to change the destination
// of a cargo.
func MakeChangeDestinationEndpoint(s booking.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(changeDestinationRequest)

		if err := s.ChangeDestination(req.ID, req.Destination); err != nil {
			return nil, err
		}

		return changeDestinationResponse{}, nil
	}
}

// List locations

type listLocationsRequest struct {
}

type listLocationsResponse struct {
	Locations []booking.Location `json:"locations"`
}

// DecodeListLocationsRequest returns a decoded request to list locations.
func DecodeListLocationsRequest(r *http.Request) (interface{}, error) {
	return listLocationsRequest{}, nil
}

// MakeListLocationsEndpoint returns a endpoint to list locations.
func MakeListLocationsEndpoint(bs booking.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		_ = request.(listLocationsRequest)
		return listLocationsResponse{Locations: bs.Locations()}, nil
	}
}

// Register incident

type registerIncidentRequest struct {
	ID             cargo.TrackingID
	Location       location.UNLocode
	Voyage         voyage.Number
	EventType      cargo.HandlingEventType
	CompletionTime time.Time
}

type registerIncidentResponse struct {
}

// DecodeRegisterIncidentRequest returns a decoded request to register an
// incident.
func DecodeRegisterIncidentRequest(r *http.Request) (interface{}, error) {
	var (
		completionTime = r.URL.Query().Get("completionTime")
		trackingID     = r.URL.Query().Get("trackingId")
		voyageNumber   = r.URL.Query().Get("voyage")
		unLocode       = r.URL.Query().Get("location")
		eventType      = r.URL.Query().Get("eventType")
	)

	millis, _ := strconv.ParseInt(completionTime, 10, 64)

	return registerIncidentRequest{
		CompletionTime: time.Unix(millis/1000, 0),
		ID:             cargo.TrackingID(trackingID),
		Voyage:         voyage.Number(voyageNumber),
		Location:       location.UNLocode(unLocode),
		EventType:      stringToEventType(eventType),
	}, nil
}

func stringToEventType(s string) cargo.HandlingEventType {
	types := map[string]cargo.HandlingEventType{
		cargo.Receive.String(): cargo.Receive,
		cargo.Load.String():    cargo.Load,
		cargo.Unload.String():  cargo.Unload,
		cargo.Customs.String(): cargo.Customs,
		cargo.Claim.String():   cargo.Claim,
	}
	return types[s]
}

// MakeRegisterIncidentEndpoint returns a endpoint to register incidents.
func MakeRegisterIncidentEndpoint(hs handling.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(registerIncidentRequest)

		if err := hs.RegisterHandlingEvent(
			req.CompletionTime, req.ID, req.Voyage, req.Location, req.EventType,
		); err != nil {
			return nil, err
		}

		return registerIncidentResponse{}, nil
	}
}

// Fetch routes

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

// DecodeFetchRoutesResponse returns a decoded request to fetch routes.
func DecodeFetchRoutesResponse(resp *http.Response) (interface{}, error) {
	var response fetchRoutesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response, nil
}

// EncodeFetchRoutesRequest encodes a fetch routes request to a HTTP request.
func EncodeFetchRoutesRequest(r *http.Request, request interface{}) error {
	req := request.(fetchRoutesRequest)

	vals := r.URL.Query()
	vals.Add("from", req.From)
	vals.Add("to", req.To)
	r.URL.RawQuery = vals.Encode()

	return nil
}
