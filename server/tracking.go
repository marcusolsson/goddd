package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"

	shipping "github.com/marcusolsson/goddd"
	"github.com/marcusolsson/goddd/tracking"
)

type trackCargoRequest struct {
	ID string
}

type trackCargoResponse struct {
	Cargo *tracking.Cargo `json:"cargo,omitempty"`
	Err   error           `json:"error,omitempty"`
}

func (r trackCargoResponse) error() error { return r.Err }

func makeTrackCargoEndpoint(ts tracking.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(trackCargoRequest)
		c, err := ts.Track(req.ID)
		return trackCargoResponse{Cargo: &c, Err: err}, nil
	}
}

// makeTrackingHandler returns a handler for the tracking service.
func makeTrackingHandler(ts tracking.Service, logger kitlog.Logger) http.Handler {
	r := mux.NewRouter()

	opts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(encodeError),
	}

	trackCargoHandler := kithttp.NewServer(
		makeTrackCargoEndpoint(ts),
		decodeTrackCargoRequest,
		encodeResponse,
		opts...,
	)

	r.Handle("/tracking/v1/cargos/{id}", trackCargoHandler).Methods("GET")
	r.Handle("/tracking/v1/docs", http.StripPrefix("/tracking/v1/docs", http.FileServer(http.Dir("tracking/docs"))))

	return r
}

func decodeTrackCargoRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, errors.New("bad route")
	}
	return trackCargoRequest{ID: id}, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

type errorer interface {
	error() error
}

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	case shipping.ErrUnknownCargo:
		w.WriteHeader(http.StatusNotFound)
	case tracking.ErrInvalidArgument:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
