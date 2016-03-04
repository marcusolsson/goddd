package booking

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/routing"
	"github.com/marcusolsson/goddd/voyage"
)

var errMissingParameters = errors.New("missing parameters")

func encodeResponse(w http.ResponseWriter, resp interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(resp)
}

func decodeBookCargoRequest(r *http.Request) (interface{}, error) {
	var body struct {
		Origin          string `json:"origin"`
		Destination     string `json:"destination"`
		ArrivalDeadline int64  `json:"arrival_deadline"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if body.Origin == "" || body.Destination == "" || body.ArrivalDeadline == 0 {
		return nil, errMissingParameters
	}

	return bookCargoRequest{
		Origin:          location.UNLocode(body.Origin),
		Destination:     location.UNLocode(body.Destination),
		ArrivalDeadline: time.Unix(body.ArrivalDeadline/1000, 0),
	}, nil
}

func decodeRequestRoutesRequest(r *http.Request) (interface{}, error) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		return nil, errMissingParameters
	}
	return requestRoutesRequest{ID: cargo.TrackingID(id)}, nil
}

func decodeAssignToRouteRequest(r *http.Request) (interface{}, error) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		return nil, errMissingParameters
	}

	var route routing.Route
	if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
		return nil, err
	}
	defer r.Body.Close()

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

	return assignToRouteRequest{
		ID:        cargo.TrackingID(id),
		Itinerary: cargo.Itinerary{Legs: legs},
	}, nil
}

func decodeChangeDestinationRequest(r *http.Request) (interface{}, error) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		return nil, errMissingParameters
	}

	var body struct {
		Destination string `json:"destination"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if body.Destination == "" {
		return nil, errMissingParameters
	}

	return changeDestinationRequest{
		ID:          cargo.TrackingID(id),
		Destination: location.UNLocode(body.Destination),
	}, nil
}

func decodeListCargosRequest(r *http.Request) (interface{}, error) {
	return listCargosRequest{}, nil
}

func decodeListLocationsRequest(r *http.Request) (interface{}, error) {
	return listLocationsRequest{}, nil
}

func decodeFetchRoutesResponse(resp *http.Response) (interface{}, error) {
	var response fetchRoutesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return response, nil
}

func encodeFetchRoutesRequest(r *http.Request, request interface{}) error {
	req := request.(fetchRoutesRequest)

	vals := r.URL.Query()
	vals.Add("from", req.From)
	vals.Add("to", req.To)
	r.URL.RawQuery = vals.Encode()

	return nil
}
