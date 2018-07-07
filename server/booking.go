package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"

	shipping "github.com/marcusolsson/goddd"
	"github.com/marcusolsson/goddd/booking"
)

type bookCargoRequest struct {
	Origin          shipping.UNLocode
	Destination     shipping.UNLocode
	ArrivalDeadline time.Time
}

type bookCargoResponse struct {
	ID  shipping.TrackingID `json:"tracking_id,omitempty"`
	Err error               `json:"error,omitempty"`
}

func (r bookCargoResponse) error() error { return r.Err }

func makeBookCargoEndpoint(s booking.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(bookCargoRequest)
		id, err := s.BookNewCargo(req.Origin, req.Destination, req.ArrivalDeadline)
		return bookCargoResponse{ID: id, Err: err}, nil
	}
}

type loadCargoRequest struct {
	ID shipping.TrackingID
}

type loadCargoResponse struct {
	Cargo *booking.Cargo `json:"cargo,omitempty"`
	Err   error          `json:"error,omitempty"`
}

func (r loadCargoResponse) error() error { return r.Err }

func makeLoadCargoEndpoint(s booking.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(loadCargoRequest)
		c, err := s.LoadCargo(req.ID)
		return loadCargoResponse{Cargo: &c, Err: err}, nil
	}
}

type requestRoutesRequest struct {
	ID shipping.TrackingID
}

type requestRoutesResponse struct {
	Routes []shipping.Itinerary `json:"routes,omitempty"`
	Err    error                `json:"error,omitempty"`
}

func (r requestRoutesResponse) error() error { return r.Err }

func makeRequestRoutesEndpoint(s booking.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(requestRoutesRequest)
		itin := s.RequestPossibleRoutesForCargo(req.ID)
		return requestRoutesResponse{Routes: itin, Err: nil}, nil
	}
}

type assignToRouteRequest struct {
	ID        shipping.TrackingID
	Itinerary shipping.Itinerary
}

type assignToRouteResponse struct {
	Err error `json:"error,omitempty"`
}

func (r assignToRouteResponse) error() error { return r.Err }

func makeAssignToRouteEndpoint(s booking.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(assignToRouteRequest)
		err := s.AssignCargoToRoute(req.ID, req.Itinerary)
		return assignToRouteResponse{Err: err}, nil
	}
}

type changeDestinationRequest struct {
	ID          shipping.TrackingID
	Destination shipping.UNLocode
}

type changeDestinationResponse struct {
	Err error `json:"error,omitempty"`
}

func (r changeDestinationResponse) error() error { return r.Err }

func makeChangeDestinationEndpoint(s booking.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(changeDestinationRequest)
		err := s.ChangeDestination(req.ID, req.Destination)
		return changeDestinationResponse{Err: err}, nil
	}
}

type listCargosRequest struct{}

type listCargosResponse struct {
	Cargos []booking.Cargo `json:"cargos,omitempty"`
	Err    error           `json:"error,omitempty"`
}

func (r listCargosResponse) error() error { return r.Err }

func makeListCargosEndpoint(s booking.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		_ = request.(listCargosRequest)
		return listCargosResponse{Cargos: s.Cargos(), Err: nil}, nil
	}
}

type listLocationsRequest struct {
}

type listLocationsResponse struct {
	Locations []booking.Location `json:"locations,omitempty"`
	Err       error              `json:"error,omitempty"`
}

func makeListLocationsEndpoint(s booking.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		_ = request.(listLocationsRequest)
		return listLocationsResponse{Locations: s.Locations(), Err: nil}, nil
	}
}

func makeBookingHandler(bs booking.Service, logger kitlog.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(encodeError),
	}

	bookCargoHandler := kithttp.NewServer(
		makeBookCargoEndpoint(bs),
		decodeBookCargoRequest,
		encodeResponse,
		opts...,
	)
	loadCargoHandler := kithttp.NewServer(
		makeLoadCargoEndpoint(bs),
		decodeLoadCargoRequest,
		encodeResponse,
		opts...,
	)
	requestRoutesHandler := kithttp.NewServer(
		makeRequestRoutesEndpoint(bs),
		decodeRequestRoutesRequest,
		encodeResponse,
		opts...,
	)
	assignToRouteHandler := kithttp.NewServer(
		makeAssignToRouteEndpoint(bs),
		decodeAssignToRouteRequest,
		encodeResponse,
		opts...,
	)
	changeDestinationHandler := kithttp.NewServer(
		makeChangeDestinationEndpoint(bs),
		decodeChangeDestinationRequest,
		encodeResponse,
		opts...,
	)
	listCargosHandler := kithttp.NewServer(
		makeListCargosEndpoint(bs),
		decodeListCargosRequest,
		encodeResponse,
		opts...,
	)
	listLocationsHandler := kithttp.NewServer(
		makeListLocationsEndpoint(bs),
		decodeListLocationsRequest,
		encodeResponse,
		opts...,
	)

	r := mux.NewRouter()

	r.Handle("/booking/v1/cargos", bookCargoHandler).Methods("POST")
	r.Handle("/booking/v1/cargos", listCargosHandler).Methods("GET")
	r.Handle("/booking/v1/cargos/{id}", loadCargoHandler).Methods("GET")
	r.Handle("/booking/v1/cargos/{id}/request_routes", requestRoutesHandler).Methods("GET")
	r.Handle("/booking/v1/cargos/{id}/assign_to_route", assignToRouteHandler).Methods("POST")
	r.Handle("/booking/v1/cargos/{id}/change_destination", changeDestinationHandler).Methods("POST")
	r.Handle("/booking/v1/locations", listLocationsHandler).Methods("GET")
	r.Handle("/booking/v1/docs", http.StripPrefix("/booking/v1/docs", http.FileServer(http.Dir("booking/docs"))))

	return r
}

var errBadRoute = errors.New("bad route")

func decodeBookCargoRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body struct {
		Origin          string    `json:"origin"`
		Destination     string    `json:"destination"`
		ArrivalDeadline time.Time `json:"arrival_deadline"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

	return bookCargoRequest{
		Origin:          shipping.UNLocode(body.Origin),
		Destination:     shipping.UNLocode(body.Destination),
		ArrivalDeadline: body.ArrivalDeadline,
	}, nil
}

func decodeLoadCargoRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, errBadRoute
	}
	return loadCargoRequest{ID: shipping.TrackingID(id)}, nil
}

func decodeRequestRoutesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, errBadRoute
	}
	return requestRoutesRequest{ID: shipping.TrackingID(id)}, nil
}

func decodeAssignToRouteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, errBadRoute
	}

	var itinerary shipping.Itinerary
	if err := json.NewDecoder(r.Body).Decode(&itinerary); err != nil {
		return nil, err
	}

	return assignToRouteRequest{
		ID:        shipping.TrackingID(id),
		Itinerary: itinerary,
	}, nil
}

func decodeChangeDestinationRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, errBadRoute
	}

	var body struct {
		Destination string `json:"destination"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

	return changeDestinationRequest{
		ID:          shipping.TrackingID(id),
		Destination: shipping.UNLocode(body.Destination),
	}, nil
}

func decodeListCargosRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return listCargosRequest{}, nil
}

func decodeListLocationsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return listLocationsRequest{}, nil
}
