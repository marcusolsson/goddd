package handling

import (
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"

	"golang.org/x/net/context"
)

// MakeHandler ...
func MakeHandler(ctx context.Context, hs Service) http.Handler {
	r := mux.NewRouter()

	r.Handle("/handling/v1/incidents", httptransport.NewServer(
		ctx,
		makeRegisterIncidentEndpoint(hs),
		decodeRegisterIncidentRequest,
		encodeResponse,
	)).Methods("POST")

	return r
}
