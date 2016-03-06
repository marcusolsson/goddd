package booking

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	httptransport "github.com/go-kit/kit/transport/http"
	"golang.org/x/net/context"

	"github.com/gorilla/mux"
	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/routing"
	"github.com/marcusolsson/goddd/voyage"
)

// MakeHandler ...
func MakeHandler(ctx context.Context, bs Service) http.Handler {
	bookCargoHandler := httptransport.NewServer(
		ctx,
		makeBookCargoEndpoint(bs),
		decodeBookCargoRequest,
		encodeResponse,
	)
	loadCargoHandler := httptransport.NewServer(
		ctx,
		makeLoadCargoEndpoint(bs),
		decodeLoadCargoRequest,
		encodeResponse,
	)
	requestRoutesHandler := httptransport.NewServer(
		ctx,
		makeRequestRoutesEndpoint(bs),
		decodeRequestRoutesRequest,
		encodeResponse,
	)
	assignToRouteHandler := httptransport.NewServer(
		ctx,
		makeAssignToRouteEndpoint(bs),
		decodeAssignToRouteRequest,
		encodeResponse,
	)
	changeDestinationHandler := httptransport.NewServer(
		ctx,
		makeChangeDestinationEndpoint(bs),
		decodeChangeDestinationRequest,
		encodeResponse,
	)
	listCargosHandler := httptransport.NewServer(
		ctx,
		makeListCargosEndpoint(bs),
		decodeListCargosRequest,
		encodeResponse,
	)
	listLocationsHandler := httptransport.NewServer(
		ctx,
		makeListLocationsEndpoint(bs),
		decodeListLocationsRequest,
		encodeResponse,
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

var errMissingParameters = errors.New("missing parameters")

func decodeBookCargoRequest(r *http.Request) (interface{}, error) {
	var body struct {
		Origin          string    `json:"origin"`
		Destination     string    `json:"destination"`
		ArrivalDeadline time.Time `json:"arrival_deadline"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if body.Origin == "" || body.Destination == "" || body.ArrivalDeadline.IsZero() {
		return nil, errMissingParameters
	}

	return bookCargoRequest{
		Origin:          location.UNLocode(body.Origin),
		Destination:     location.UNLocode(body.Destination),
		ArrivalDeadline: body.ArrivalDeadline,
	}, nil
}

func decodeLoadCargoRequest(r *http.Request) (interface{}, error) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		return nil, errors.New("missing parameters")
	}
	return loadCargoRequest{ID: cargo.TrackingID(id)}, nil
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

func encodeResponse(w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(w, e.error())
		return nil
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

type errorer interface {
	error() error
}

// encode errors from business-logic
func encodeError(w http.ResponseWriter, err error) {
	switch err {
	case nil:
		w.WriteHeader(http.StatusOK)
	case cargo.ErrUnknown:
		w.WriteHeader(http.StatusNotFound)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
