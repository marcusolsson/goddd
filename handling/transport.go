package handling

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
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
	var (
		completionTime int64
		trackingID     string
		voyageNumber   string
		unLocode       string
		eventType      string
	)

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		var data struct {
			CompletionTime int64
			TrackingID     string
			VoyageNumber   string
			Location       string
			EventType      string
		}

		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			return nil, err
		}
		defer r.Body.Close()

		completionTime = data.CompletionTime
		trackingID = data.TrackingID
		voyageNumber = data.VoyageNumber
		unLocode = data.Location
		eventType = data.EventType
	} else {
		trackingID = getParam(r, "tracking_id")
		voyageNumber = getParam(r, "voyage")
		unLocode = getParam(r, "location")
		eventType = getParam(r, "event_type")

		millis := getParam(r, "completion_time")
		completionTime, _ = strconv.ParseInt(millis, 10, 64)
	}

	return registerIncidentRequest{
		CompletionTime: time.Unix(completionTime/1000, 0),
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

func getParam(r *http.Request, key string) string {
	if val := r.FormValue(key); val != "" {
		return val
	}
	if val := r.URL.Query().Get(key); val != "" {
		return val
	}

	return ""
}
