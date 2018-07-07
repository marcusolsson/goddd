package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	kitlog "github.com/go-kit/kit/log"

	shipping "github.com/marcusolsson/goddd"
	"github.com/marcusolsson/goddd/handling"
)

type handlingHandler struct {
	s handling.Service

	logger kitlog.Logger
}

func (h *handlingHandler) router() chi.Router {
	r := chi.NewRouter()
	r.Post("/incidents", h.registerIncident)
	r.Method("GET", "/docs", http.StripPrefix("/handling/v1/docs", http.FileServer(http.Dir("handling/docs"))))
	return r
}

func (h *handlingHandler) registerIncident(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var request struct {
		CompletionTime time.Time `json:"completion_time"`
		TrackingID     string    `json:"tracking_id"`
		VoyageNumber   string    `json:"voyage"`
		Location       string    `json:"location"`
		EventType      string    `json:"event_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Log("error", err)
		encodeError(ctx, err, w)
		return
	}

	err := h.s.RegisterHandlingEvent(
		request.CompletionTime,
		shipping.TrackingID(request.TrackingID),
		shipping.VoyageNumber(request.VoyageNumber),
		shipping.UNLocode(request.Location),
		stringToEventType(request.EventType),
	)
	if err != nil {
		encodeError(ctx, err, w)
		return
	}
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
