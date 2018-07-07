package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"

	shipping "github.com/marcusolsson/goddd"
	"github.com/marcusolsson/goddd/handling"
)

type registerIncidentRequest struct {
	ID             shipping.TrackingID
	Location       shipping.UNLocode
	Voyage         shipping.VoyageNumber
	EventType      shipping.HandlingEventType
	CompletionTime time.Time
}

type registerIncidentResponse struct {
	Err error `json:"error,omitempty"`
}

func (r registerIncidentResponse) error() error { return r.Err }

func makeRegisterIncidentEndpoint(hs handling.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(registerIncidentRequest)
		err := hs.RegisterHandlingEvent(req.CompletionTime, req.ID, req.Voyage, req.Location, req.EventType)
		return registerIncidentResponse{Err: err}, nil
	}
}

func makeHandlingHandler(hs handling.Service, logger kitlog.Logger) http.Handler {
	r := mux.NewRouter()

	opts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(encodeError),
	}

	registerIncidentHandler := kithttp.NewServer(
		makeRegisterIncidentEndpoint(hs),
		decodeRegisterIncidentRequest,
		encodeResponse,
		opts...,
	)

	r.Handle("/handling/v1/incidents", registerIncidentHandler).Methods("POST")
	r.Handle("/handling/v1/docs", http.StripPrefix("/handling/v1/docs", http.FileServer(http.Dir("handling/docs"))))

	return r
}

func decodeRegisterIncidentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body struct {
		CompletionTime time.Time `json:"completion_time"`
		TrackingID     string    `json:"tracking_id"`
		VoyageNumber   string    `json:"voyage"`
		Location       string    `json:"location"`
		EventType      string    `json:"event_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

	return registerIncidentRequest{
		CompletionTime: body.CompletionTime,
		ID:             shipping.TrackingID(body.TrackingID),
		Voyage:         shipping.VoyageNumber(body.VoyageNumber),
		Location:       shipping.UNLocode(body.Location),
		EventType:      stringToEventType(body.EventType),
	}, nil
}

func stringToEventType(s string) shipping.HandlingEventType {
	types := map[string]shipping.HandlingEventType{
		shipping.Receive.String(): shipping.Receive,
		shipping.Load.String():    shipping.Load,
		shipping.Unload.String():  shipping.Unload,
		shipping.Customs.String(): shipping.Customs,
		shipping.Claim.String():   shipping.Claim,
	}
	return types[s]
}
