package main

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
	"github.com/marcusolsson/goddd/voyage"
)

func encodeResponse(w http.ResponseWriter, resp interface{}) error {
	return json.NewEncoder(w).Encode(resp)
}

// List cargos

type listCargosRequest struct{}

type listCargosResponse struct {
	Cargos []cargoView `json:"cargos"`
}

func makeListCargosEndpoint(cs CargoProjectionService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return listCargosResponse{Cargos: cs.FindAll()}, nil
	}
}

func decodeListCargosRequest(r *http.Request) (interface{}, error) {
	return listCargosRequest{}, nil
}

// Find cargo

type findCargoRequest struct {
	ID string
}

type findCargoResponse struct {
	Cargo cargoView `json:"cargo"`
}

func makeFindCargoEndpoint(cs CargoProjectionService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(findCargoRequest)

		c, err := cs.Find(req.ID)
		if err != nil {
			return nil, err
		}

		return findCargoResponse{Cargo: c}, nil
	}
}

func decodeFindCargoRequest(r *http.Request) (interface{}, error) {
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

func makeBookCargoEndpoint(s booking.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(bookCargoRequest)

		id, err := s.BookNewCargo(req.Origin, req.Destination, req.ArrivalDeadline)
		if err != nil {
			return nil, err
		}

		return bookCargoResponse{ID: id}, nil
	}
}

func decodeBookCargoRequest(r *http.Request) (interface{}, error) {
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
	Routes []routeView `json:"routes"`
}

func makeRequestRoutesEndpoint(s booking.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(requestRoutesRequest)

		itineraries := s.RequestPossibleRoutesForCargo(req.ID)

		result := []routeView{}
		for _, itin := range itineraries {
			var legs []legView
			for _, leg := range itin.Legs {
				legs = append(legs, legView{
					VoyageNumber: string(leg.VoyageNumber),
					From:         string(leg.LoadLocation),
					To:           string(leg.UnloadLocation),
					LoadTime:     leg.LoadTime,
					UnloadTime:   leg.UnloadTime,
				})
			}
			result = append(result, routeView{Legs: legs})
		}

		return requestRoutesResponse{Routes: result}, nil
	}
}

func decodeRequestRoutesRequest(r *http.Request) (interface{}, error) {
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

func makeAssignToRouteEndpoint(s booking.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(assignToRouteRequest)

		if err := s.AssignCargoToRoute(req.ID, req.Itinerary); err != nil {
			return nil, err
		}

		return assignToRouteResponse{}, nil
	}
}

func decodeAssignToRouteRequest(r *http.Request) (interface{}, error) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var route routeView
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

func decodeChangeDestinationRequest(r *http.Request) (interface{}, error) {
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

func makeChangeDestinationEndpoint(s booking.Service) endpoint.Endpoint {
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
	Locations []location.Location
}

func decodeListLocationsRequest(r *http.Request) (interface{}, error) {
	return listLocationsRequest{}, nil
}

func makeListLocationsEndpoint(ls LocationProjectionService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		_ = request.(listLocationsRequest)

		locations := ls.FindAll()

		return listLocationsResponse{Locations: locations}, nil
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

func decodeRegisterIncidentRequest(r *http.Request) (interface{}, error) {
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

func makeRegisterIncidentEndpoint(hs handling.Service) endpoint.Endpoint {
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
