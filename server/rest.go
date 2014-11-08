package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/hariharan-uno/cors"
	"gopkg.in/unrolled/render.v1"

	"github.com/marcusolsson/goddd/booking"
	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/handling"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/repository"
	"github.com/marcusolsson/goddd/routing"
)

type API struct {
	Renderer *render.Render
	Booking  booking.Facade
	Handling handling.Facade
}

func NewAPI() *API {

	var (
		cargoRepository            = repository.NewCargo()
		locationRepository         = repository.NewLocation()
		voyageRepository           = repository.NewVoyage()
		handlingEventRepository    = repository.NewHandlingEvent()
		routingService             = routing.NewService(locationRepository)
		bookingService             = booking.NewService(cargoRepository, locationRepository, routingService)
		bookingServiceFacade       = booking.NewFacade(bookingService)
		handlingEventServiceFacade = handling.NewFacade(cargoRepository, locationRepository, voyageRepository, handlingEventRepository)
	)

	storeTestData(cargoRepository)

	return &API{
		Renderer: render.New(render.Options{IndentJSON: true}),
		Booking:  bookingServiceFacade,
		Handling: handlingEventServiceFacade,
	}
}

func RegisterHandlers() {

	a := NewAPI()

	router := mux.NewRouter()

	router.HandleFunc("/cargos", a.CargosHandler).Methods("GET")
	router.HandleFunc("/cargos", a.BookCargoHandler).Methods("POST")
	router.HandleFunc("/cargos/{id}", a.CargoHandler).Methods("GET")
	router.HandleFunc("/cargos/{id}/request_routes", a.RequestRoutesHandler).Methods("GET")
	router.HandleFunc("/cargos/{id}/assign_to_route", a.AssignToRouteHandler).Methods("POST")
	router.HandleFunc("/cargos/{id}/change_destination", a.ChangeDestinationHandler).Methods("POST")
	router.HandleFunc("/locations", a.LocationsHandler).Methods("GET")
	router.HandleFunc("/incidents", a.RegisterIncidentHandler).Methods("POST")

	n := negroni.Classic()

	options := cors.Options{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type"},
	}

	n.Use(negroni.HandlerFunc(options.Allow))
	n.UseHandler(router)

	http.Handle("/", n)
}

var (
	ResourceNotFound = map[string]interface{}{"error": "resource not found"}
	InvalidInput     = map[string]interface{}{"error": "invalid input"}
)

func (a *API) CargosHandler(w http.ResponseWriter, req *http.Request) {
	a.Renderer.JSON(w, http.StatusOK, a.Booking.ListAllCargos())
}

func (a *API) CargoHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	c, err := a.Booking.LoadCargoForRouting(vars["id"])

	if err != nil {
		a.Renderer.JSON(w, http.StatusNotFound, ResourceNotFound)
		return
	}

	a.Renderer.JSON(w, http.StatusOK, c)
}

func (a *API) RequestRoutesHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	a.Renderer.JSON(w, http.StatusOK, a.Booking.RequestRoutesForCargo(vars["id"]))
}

func (a *API) AssignToRouteHandler(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	candidate := booking.RouteCandidateDTO{}
	if err := decoder.Decode(&candidate); err != nil {
		a.Renderer.JSON(w, http.StatusBadRequest, InvalidInput)
	}

	vars := mux.Vars(req)
	if err := a.Booking.AssignCargoToRoute(vars["id"], candidate); err != nil {
		a.Renderer.JSON(w, http.StatusBadRequest, InvalidInput)
		return
	}

	a.Renderer.JSON(w, http.StatusOK, map[string]interface{}{})
}

func (a *API) ChangeDestinationHandler(w http.ResponseWriter, req *http.Request) {
	v := queryParams(req.URL.Query())
	found, missing := v.validateQueryParams("destination")

	if len(missing) > 0 {
		a.Renderer.JSON(w, http.StatusBadRequest, map[string]interface{}{"error": "missing parameter", "missing": missing})
		return
	}

	vars := mux.Vars(req)
	if err := a.Booking.ChangeDestination(vars["id"], fmt.Sprintf("%s", found["destination"])); err != nil {
		a.Renderer.JSON(w, http.StatusBadRequest, InvalidInput)
		return
	}

	a.Renderer.JSON(w, http.StatusOK, map[string]interface{}{})
}

func (a *API) BookCargoHandler(w http.ResponseWriter, req *http.Request) {
	v := queryParams(req.URL.Query())
	found, missing := v.validateQueryParams("origin", "destination", "arrivalDeadline")

	if len(missing) > 0 {
		a.Renderer.JSON(w, http.StatusBadRequest, map[string]interface{}{"error": "missing parameter", "missing": missing})
		return
	}

	var (
		origin          = found["origin"].(string)
		destination     = found["destination"].(string)
		arrivalDeadline = found["arrivalDeadline"].(string)
	)

	trackingID, err := a.Booking.BookNewCargo(origin, destination, arrivalDeadline)

	if err != nil {
		a.Renderer.JSON(w, http.StatusBadRequest, InvalidInput)
		return
	}

	c, err := a.Booking.LoadCargoForRouting(trackingID)

	if err != nil {
		a.Renderer.JSON(w, http.StatusNotFound, ResourceNotFound)
		return
	}

	a.Renderer.JSON(w, http.StatusOK, c)
}

func (a *API) LocationsHandler(w http.ResponseWriter, req *http.Request) {
	a.Renderer.JSON(w, http.StatusOK, a.Booking.ListShippingLocations())
}

func (a *API) RegisterIncidentHandler(w http.ResponseWriter, req *http.Request) {
	v := queryParams(req.URL.Query())
	found, missing := v.validateQueryParams("completionTime", "trackingId", "voyage", "location", "eventType")

	if len(missing) > 0 {
		a.Renderer.JSON(w, http.StatusBadRequest, map[string]interface{}{"error": "missing parameter", "missing": missing})
		return
	}

	var (
		completionTime = found["completionTime"].(string)
		trackingID     = found["trackingId"].(string)
		voyageNumber   = found["voyage"].(string)
		location       = found["location"].(string)
		eventType      = found["eventType"].(string)
	)

	if err := a.Handling.RegisterHandlingEvent(completionTime, trackingID, voyageNumber, location, eventType); err != nil {
		a.Renderer.JSON(w, http.StatusBadRequest, InvalidInput)
		return
	}
}

func storeTestData(r cargo.Repository) {
	test1 := cargo.New("FTL456", cargo.RouteSpecification{
		Origin:      location.AUMEL,
		Destination: location.SESTO,
	})
	r.Store(*test1)

	test2 := cargo.New("ABC123", cargo.RouteSpecification{
		Origin:      location.SESTO,
		Destination: location.CNHKG,
	})
	r.Store(*test2)
}

type queryParams url.Values

type jsonObject map[string]interface{}

func (p queryParams) validateQueryParams(params ...string) (found jsonObject, missing []string) {
	found = make(jsonObject)
	missing = make([]string, 0)

	for _, param := range params {
		s := url.Values(p).Get(param)
		if len(s) > 0 {
			found[param] = s
		} else {
			missing = append(missing, param)
		}
	}
	return found, missing
}
