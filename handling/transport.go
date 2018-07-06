package handling

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"

	shipping "github.com/marcusolsson/goddd"
)

// MakeHandler returns a handler for the handling service.
func MakeHandler(hs Service, logger kitlog.Logger) http.Handler {
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
	case ErrInvalidArgument:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
