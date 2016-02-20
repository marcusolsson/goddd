package tracking

import (
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
)

// MakeHandler ...
func MakeHandler(ctx context.Context, ts Service) http.Handler {
	r := mux.NewRouter()

	r.Handle("/tracking/v1/cargos/{id}", httptransport.NewServer(
		ctx,
		makeFindCargoEndpoint(ts),
		decodeFindCargoRequest,
		encodeFindCargoResponse,
	)).Methods("GET")

	return r
}
