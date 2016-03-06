package tracking

import (
	"encoding/json"
	"errors"
	"net/http"

	"golang.org/x/net/context"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/marcusolsson/goddd/cargo"
)

// MakeHandler ...
func MakeHandler(ctx context.Context, ts Service) http.Handler {
	r := mux.NewRouter()

	findCargoHandler := httptransport.NewServer(
		ctx,
		makeFindCargoEndpoint(ts),
		decodeFindCargoRequest,
		encodeFindCargoResponse,
	)

	r.Handle("/tracking/v1/cargos/{id}", findCargoHandler).Methods("GET")
	r.Handle("/tracking/v1/docs", http.StripPrefix("/tracking/v1/docs", http.FileServer(http.Dir("tracking/docs"))))

	return r
}

func decodeFindCargoRequest(r *http.Request) (interface{}, error) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		return nil, errors.New("missing parameters")
	}
	return findCargoRequest{ID: id}, nil
}

func encodeFindCargoResponse(w http.ResponseWriter, response interface{}) error {
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
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
