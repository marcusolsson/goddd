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
