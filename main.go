package main

import (
	"net/http"
	"os"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"

	"github.com/marcusolsson/goddd/booking"
	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/handling"
	"github.com/marcusolsson/goddd/inspection"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/middleware"
	"github.com/marcusolsson/goddd/repository"
	"github.com/marcusolsson/goddd/routing"
	"github.com/marcusolsson/goddd/tracking"
)

var (
	defaultPort              = "8080"
	defaultRoutingServiceURL = "http://localhost:7878"
)

func main() {
	// Configure repositories.
	var (
		cargos         = repository.NewCargo()
		locations      = repository.NewLocation()
		voyages        = repository.NewVoyage()
		handlingEvents = repository.NewHandlingEvent()
	)

	// Configure some questionable dependencies.
	var (
		handlingEventFactory = cargo.HandlingEventFactory{
			CargoRepository:    cargos,
			VoyageRepository:   voyages,
			LocationRepository: locations,
		}
		handlingEventHandler = handling.NewEventHandler(
			inspection.NewService(cargos, handlingEvents, nil),
		)
	)

	// Facilitate testing by adding some cargos.
	storeTestData(cargos)

	ctx := context.Background()
	logger := log.NewLogfmtLogger(os.Stdout)

	// Configure the routing service which will serve as a proxy.
	var rs routing.Service
	rs = middleware.ProxyingRoutingMiddleware(routingServiceURL(), ctx)(rs)

	// Create handlers for all booking endpoints.
	var bs booking.Service
	bs = booking.NewService(cargos, locations, handlingEvents, rs)
	bs = middleware.LoggingBookingService{logger, bs}

	bookCargoHandler := httptransport.NewServer(
		ctx,
		middleware.MakeBookCargoEndpoint(bs),
		middleware.DecodeBookCargoRequest,
		middleware.EncodeResponse,
	)
	requestRoutesHandler := httptransport.NewServer(
		ctx,
		middleware.MakeRequestRoutesEndpoint(bs),
		middleware.DecodeRequestRoutesRequest,
		middleware.EncodeResponse,
	)
	assignToRouteHandler := httptransport.NewServer(
		ctx,
		middleware.MakeAssignToRouteEndpoint(bs),
		middleware.DecodeAssignToRouteRequest,
		middleware.EncodeResponse,
	)
	changeDestinationHandler := httptransport.NewServer(
		ctx,
		middleware.MakeChangeDestinationEndpoint(bs),
		middleware.DecodeChangeDestinationRequest,
		middleware.EncodeResponse,
	)
	listCargosHandler := httptransport.NewServer(
		ctx,
		middleware.MakeListCargosEndpoint(bs),
		middleware.DecodeListCargosRequest,
		middleware.EncodeResponse,
	)
	listLocationsHandler := httptransport.NewServer(
		ctx,
		middleware.MakeListLocationsEndpoint(bs),
		middleware.DecodeListLocationsRequest,
		middleware.EncodeResponse,
	)

	var ts tracking.Service
	ts = tracking.NewService(cargos, handlingEvents)
	ts = middleware.LoggingTrackingService{logger, ts}

	findCargoHandler := httptransport.NewServer(
		ctx,
		middleware.MakeFindCargoEndpoint(ts),
		middleware.DecodeFindCargoRequest,
		middleware.EncodeFindCargoResponse,
	)

	// Create handler for the handling endpoint used for registering incidents.
	var hs handling.Service
	hs = handling.NewService(handlingEvents, handlingEventFactory, handlingEventHandler)

	registerIncidentHandler := httptransport.NewServer(
		ctx,
		middleware.MakeRegisterIncidentEndpoint(hs),
		middleware.DecodeRegisterIncidentRequest,
		middleware.EncodeResponse,
	)

	// Register all handlers and start serving ...
	r := mux.NewRouter()

	r.Handle("/cargos", bookCargoHandler).Methods("POST")
	r.Handle("/cargos", listCargosHandler).Methods("GET")
	r.Handle("/cargos/{id}", findCargoHandler).Methods("GET")
	r.Handle("/cargos/{id}/request_routes", requestRoutesHandler).Methods("GET")
	r.Handle("/cargos/{id}/assign_to_route", assignToRouteHandler).Methods("POST")
	r.Handle("/cargos/{id}/change_destination", changeDestinationHandler).Methods("POST")
	r.Handle("/locations", listLocationsHandler).Methods("GET")
	r.Handle("/incidents", registerIncidentHandler).Methods("POST")

	r.Handle("/", http.RedirectHandler("/docs/", http.StatusMovedPermanently))
	r.Handle("/docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir("docs"))))

	http.Handle("/", accessControl(r))

	addr := ":" + port()

	_ = logger.Log("msg", "HTTP", "addr", addr)
	_ = logger.Log("err", http.ListenAndServe(addr, nil))
}

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}

func port() string {
	port := os.Getenv("PORT")
	if port == "" {
		return defaultPort
	}
	return port
}

func routingServiceURL() string {
	url := os.Getenv("ROUTINGSERVICE_URL")
	if url == "" {
		return defaultRoutingServiceURL
	}
	return url
}

func storeTestData(r cargo.Repository) {
	test1 := cargo.New("FTL456", cargo.RouteSpecification{
		Origin:      location.AUMEL,
		Destination: location.SESTO,
	})
	_ = r.Store(*test1)

	test2 := cargo.New("ABC123", cargo.RouteSpecification{
		Origin:      location.SESTO,
		Destination: location.CNHKG,
	})
	_ = r.Store(*test2)
}
