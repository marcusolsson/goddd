package booking

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/routing"
	"github.com/marcusolsson/goddd/voyage"
)

func encodeResponse(w http.ResponseWriter, resp interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(resp)
}

func decodeBookCargoRequest(r *http.Request) (interface{}, error) {
	var (
		origin          string
		destination     string
		arrivalDeadline int64
	)

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		var data struct {
			Origin          string `json:"origin"`
			Destination     string `json:"destination"`
			ArrivalDeadline int64  `json:"arrival_deadline"`
		}

		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			return nil, err
		}

		origin = data.Origin
		destination = data.Destination
		arrivalDeadline = data.ArrivalDeadline
	} else {
		origin = getParam(r, "origin")
		destination = getParam(r, "destination")

		millis := getParam(r, "arrival_deadline")
		arrivalDeadline, _ = strconv.ParseInt(millis, 10, 64)
	}

	if origin == "" || destination == "" || arrivalDeadline == 0 {
		return nil, errors.New("missing parameters")
	}

	return bookCargoRequest{
		Origin:          location.UNLocode(origin),
		Destination:     location.UNLocode(destination),
		ArrivalDeadline: time.Unix(arrivalDeadline/1000, 0),
	}, nil
}

func decodeRequestRoutesRequest(r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)

	id := vars["id"]
	if id == "" {
		return nil, errors.New("missing parameters")
	}

	return requestRoutesRequest{ID: cargo.TrackingID(id)}, nil
}

func decodeAssignToRouteRequest(r *http.Request) (interface{}, error) {
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

func decodeChangeDestinationRequest(r *http.Request) (interface{}, error) {
	id := mux.Vars(r)["id"]

	var destination string

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		var data struct {
			Destination string `json:"destination"`
		}

		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			return nil, err
		}

		destination = data.Destination
	} else {
		destination = getParam(r, "destination")
	}

	if id == "" || destination == "" {
		return nil, errors.New("missing parameters")
	}

	return changeDestinationRequest{
		ID:          cargo.TrackingID(id),
		Destination: location.UNLocode(destination),
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

func getParam(r *http.Request, key string) string {
	if val := r.FormValue(key); val != "" {
		return val
	}
	if val := r.URL.Query().Get(key); val != "" {
		return val
	}

	return ""
}
