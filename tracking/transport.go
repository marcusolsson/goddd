package tracking

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/marcusolsson/goddd/cargo"
)

func decodeFindCargoRequest(r *http.Request) (interface{}, error) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		return nil, errors.New("missing parameters")
	}
	return findCargoRequest{ID: id}, nil
}

func encodeFindCargoResponse(w http.ResponseWriter, response interface{}) error {
	resp := response.(findCargoResponse)

	if resp.Err == cargo.ErrUnknown.Error() {
		w.WriteHeader(http.StatusNotFound)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(resp)
}
