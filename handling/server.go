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

	registerIncidentHandler := httptransport.NewServer(
		ctx,
		makeRegisterIncidentEndpoint(hs),
		decodeRegisterIncidentRequest,
		encodeResponse,
	)

	r.Handle("/handling/v1/incidents", registerIncidentHandler).Methods("POST")
	r.Handle("/handling/v1/docs", http.StripPrefix("/handling/v1/docs", http.FileServer(http.Dir("handling/docs"))))

	return r
}
