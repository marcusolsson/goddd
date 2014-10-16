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

	"github.com/marcusolsson/goddd/application"
	"github.com/marcusolsson/goddd/domain/cargo"
	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/infrastructure"
	"github.com/marcusolsson/goddd/interfaces"
)

type Api struct {
	Renderer *render.Render
	Facade   interfaces.BookingServiceFacade
}

func NewApi() *Api {

	var (
		cargoRepository         = infrastructure.NewInMemCargoRepository()
		locationRepository      = infrastructure.NewLocationRepositoryMongoDB()
		handlingEventRepository = infrastructure.NewInMemHandlingEventRepository()
		routingService          = infrastructure.NewExternalRoutingService(locationRepository)
		bookingService          = application.NewBookingService(cargoRepository, locationRepository, routingService)
		bookingServiceFacade    = interfaces.NewBookingServiceFacade(cargoRepository, locationRepository, handlingEventRepository, bookingService)
	)

	storeTestData(cargoRepository)

	return &Api{
		Renderer: render.New(render.Options{IndentJSON: true}),
		Facade:   bookingServiceFacade,
	}
}

func RegisterHandlers() {

	a := NewApi()

	router := mux.NewRouter()

	router.HandleFunc("/cargos", a.CargosHandler).Methods("GET")
	router.HandleFunc("/cargos", a.BookCargoHandler).Methods("POST")
	router.HandleFunc("/cargos/{id}", a.CargoHandler).Methods("GET")
	router.HandleFunc("/cargos/{id}/request_routes", a.RequestRoutesHandler).Methods("GET")
	router.HandleFunc("/cargos/{id}/assign_to_route", a.AssignToRouteHandler).Methods("POST")
	router.HandleFunc("/cargos/{id}/change_destination", a.ChangeDestinationHandler).Methods("POST")
	router.HandleFunc("/locations", a.LocationsHandler).Methods("GET")

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

func (a *Api) CargosHandler(w http.ResponseWriter, req *http.Request) {
	a.Renderer.JSON(w, http.StatusOK, a.Facade.ListAllCargos())
}

func (a *Api) CargoHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	c, err := a.Facade.LoadCargoForRouting(vars["id"])

	if err != nil {
		a.Renderer.JSON(w, http.StatusNotFound, ResourceNotFound)
		return
	}

	a.Renderer.JSON(w, http.StatusOK, c)
}

func (a *Api) RequestRoutesHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	a.Renderer.JSON(w, http.StatusOK, a.Facade.RequestRoutesForCargo(vars["id"]))
}

func (a *Api) AssignToRouteHandler(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	candidate := interfaces.RouteCandidateDTO{}
	if err := decoder.Decode(&candidate); err != nil {
		a.Renderer.JSON(w, http.StatusBadRequest, InvalidInput)
	}

	vars := mux.Vars(req)
	if err := a.Facade.AssignCargoToRoute(vars["id"], candidate); err != nil {
		a.Renderer.JSON(w, http.StatusBadRequest, InvalidInput)
		return
	}

	a.Renderer.JSON(w, http.StatusOK, map[string]interface{}{})
}

func (a *Api) ChangeDestinationHandler(w http.ResponseWriter, req *http.Request) {
	v := queryParams(req.URL.Query())
	found, missing := v.validateQueryParams("destination")

	if len(missing) > 0 {
		a.Renderer.JSON(w, http.StatusBadRequest, map[string]interface{}{"error": "missing parameter", "missing": missing})
		return
	}

	vars := mux.Vars(req)
	if err := a.Facade.ChangeDestination(vars["id"], fmt.Sprintf("%s", found["destination"])); err != nil {
		a.Renderer.JSON(w, http.StatusBadRequest, InvalidInput)
		return
	}

	a.Renderer.JSON(w, http.StatusOK, map[string]interface{}{})
}

func (a *Api) BookCargoHandler(w http.ResponseWriter, req *http.Request) {
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

	trackingId, err := a.Facade.BookNewCargo(origin, destination, arrivalDeadline)

	if err != nil {
		a.Renderer.JSON(w, http.StatusBadRequest, InvalidInput)
		return
	}

	c, err := a.Facade.LoadCargoForRouting(trackingId)

	if err != nil {
		a.Renderer.JSON(w, http.StatusNotFound, ResourceNotFound)
		return
	}

	a.Renderer.JSON(w, http.StatusOK, c)
}

func (a *Api) LocationsHandler(w http.ResponseWriter, req *http.Request) {
	a.Renderer.JSON(w, http.StatusOK, a.Facade.ListShippingLocations())
}

func storeTestData(r cargo.Repository) {
	test1 := cargo.NewCargo("FTL456", cargo.RouteSpecification{
		Origin:      location.AUMEL,
		Destination: location.SESTO,
	})
	r.Store(*test1)

	test2 := cargo.NewCargo("ABC123", cargo.RouteSpecification{
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
