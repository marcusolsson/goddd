package handling

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/voyage"
)

func encodeResponse(w http.ResponseWriter, resp interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(resp)
}

func decodeRegisterIncidentRequest(r *http.Request) (interface{}, error) {
	var body struct {
		CompletionTime int64
		TrackingID     string
		VoyageNumber   string
		Location       string
		EventType      string
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return registerIncidentRequest{
		CompletionTime: time.Unix(body.CompletionTime/1000, 0),
		ID:             cargo.TrackingID(body.TrackingID),
		Voyage:         voyage.Number(body.VoyageNumber),
		Location:       location.UNLocode(body.Location),
		EventType:      stringToEventType(body.EventType),
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
