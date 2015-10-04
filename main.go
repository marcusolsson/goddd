package main

import (
	"log"
	"net/http"
	"os"

	"golang.org/x/net/context"

	"github.com/gorilla/mux"
	"github.com/marcusolsson/goddd/booking"
	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/handling"
	"github.com/marcusolsson/goddd/inspection"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/repository"
	"github.com/marcusolsson/goddd/routing"

	httptransport "github.com/go-kit/kit/transport/http"
)

var (
	defaultPort              = "8080"
	defaultRoutingServiceURL = "http://localhost:7878"
)

func main() {
	// Configure repositories.
	var (
		cargoRepository         = repository.NewCargo()
		locationRepository      = repository.NewLocation()
		voyageRepository        = repository.NewVoyage()
		handlingEventRepository = repository.NewHandlingEvent()
	)

	// Configure some questionable dependencies.
	var (
		handlingEventFactory = cargo.HandlingEventFactory{
			CargoRepository:    cargoRepository,
			VoyageRepository:   voyageRepository,
			LocationRepository: locationRepository,
		}
		handlingEventHandler = handling.NewEventHandler(
			inspection.NewService(cargoRepository, handlingEventRepository, nil),
		)
	)

	// Facilitate testing by adding some cargos.
	storeTestData(cargoRepository)

	ctx := context.Background()

	// Configure the routing service which will serve as a proxy.
	var rs routing.Service
	rs = proxyingMiddleware(routingServiceURL(), ctx)(rs)
	// Create handlers for all booking endpoint.
	var bs booking.Service
	bs = booking.NewService(cargoRepository, locationRepository, rs)

	bookCargoHandler := httptransport.NewServer(
		ctx,
		makeBookCargoEndpoint(bs),
		decodeBookCargoRequest,
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

	// Create handlers for the cargo endpoints.
	var cs CargoProjectionService
	cs = &cargoProjectionService{
		cargoRepository:         cargoRepository,
		handlingEventRepository: handlingEventRepository,
	}

	listCargosHandler := httptransport.NewServer(
		ctx,
		makeListCargosEndpoint(cs),
		decodeListCargosRequest,
		encodeResponse,
	)
	findCargoHandler := httptransport.NewServer(
		ctx,
		makeFindCargoEndpoint(cs),
		decodeFindCargoRequest,
		encodeResponse,
	)

	// Create handler for the location endpoint.
	var ls LocationProjectionService
	ls = &locationProjectionService{
		repository: locationRepository,
	}

	listLocationsHandler := httptransport.NewServer(
		ctx,
		makeListLocationsEndpoint(ls),
		decodeListLocationsRequest,
		encodeResponse,
	)

	// Create handler for the handling endpoint used for registering incidents.
	var hs handling.Service
	hs = handling.NewService(handlingEventRepository, handlingEventFactory, handlingEventHandler)

	registerIncidentHandler := httptransport.NewServer(
		ctx,
		makeRegisterIncidentEndpoint(hs),
		decodeRegisterIncidentRequest,
		encodeResponse,
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

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":"+port(), nil))
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
