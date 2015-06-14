package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/hariharan-uno/cors"
	"github.com/marcusolsson/goddd/booking"
	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/handling"
	"github.com/marcusolsson/goddd/inspection"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/repository"
	"github.com/marcusolsson/goddd/routing"
	"github.com/marcusolsson/goddd/voyage"
)

var (
	defaultPort              = "8080"
	defaultRoutingServiceURL = "http://ddd-pathfinder.herokuapp.com"
)

type context struct {
	cargoRepository         cargo.Repository
	locationRepository      location.Repository
	handlingEventRepository cargo.HandlingEventRepository
	bookingService          booking.Service
	handlingEventService    handling.Service
}

type ctxHandler func(w http.ResponseWriter, r *http.Request, ctx *context)

func handlerFunc(ctx *context, fn ctxHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, ctx)
	})
}

func main() {
	var (
		cargoRepository         = repository.NewCargo()
		locationRepository      = repository.NewLocation()
		voyageRepository        = repository.NewVoyage()
		handlingEventRepository = repository.NewHandlingEvent()
		routingService          = routing.NewService(defaultRoutingServiceURL)
		bookingService          = booking.NewService(cargoRepository, locationRepository, routingService)
	)

	var (
		handlingEventFactory = cargo.HandlingEventFactory{
			CargoRepository:    cargoRepository,
			VoyageRepository:   voyageRepository,
			LocationRepository: locationRepository,
		}
		handlingEventHandler = handling.NewEventHandler(
			inspection.NewService(cargoRepository, handlingEventRepository, nil),
		)
		handlingEventService = handling.NewService(handlingEventRepository, handlingEventFactory, handlingEventHandler)
	)

	ctx := &context{
		cargoRepository:         cargoRepository,
		locationRepository:      locationRepository,
		handlingEventRepository: handlingEventRepository,
		bookingService:          bookingService,
		handlingEventService:    handlingEventService,
	}

	storeTestData(cargoRepository)

	r := mux.NewRouter()

	r.Handle("/cargos", handlerFunc(ctx, cargosHandler)).Methods("GET")
	r.Handle("/cargos", handlerFunc(ctx, bookCargoHandler)).Methods("POST")
	r.Handle("/cargos/{id}", handlerFunc(ctx, cargoHandler)).Methods("GET")
	r.Handle("/cargos/{id}/request_routes", handlerFunc(ctx, requestRoutesHandler)).Methods("GET")
	r.Handle("/cargos/{id}/assign_to_route", handlerFunc(ctx, assignToRouteHandler)).Methods("POST")
	r.Handle("/cargos/{id}/change_destination", handlerFunc(ctx, changeDestinationHandler)).Methods("POST")
	r.Handle("/locations", handlerFunc(ctx, locationsHandler)).Methods("GET")
	r.Handle("/incidents", handlerFunc(ctx, registerIncidentHandler)).Methods("POST")

	n := negroni.Classic()

	options := cors.Options{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type"},
	}

	n.Use(negroni.HandlerFunc(options.Allow))
	n.UseHandler(r)

	http.Handle("/", n)

	http.ListenAndServe(":"+port(), nil)
}

func port() string {
	p := os.Getenv("PORT")
	if p == "" {
		return defaultPort
	}
	return p
}

func cargosHandler(w http.ResponseWriter, req *http.Request, ctx *context) {
	if req.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	cargos := ctx.cargoRepository.FindAll()

	result := make([]cargoView, len(cargos))
	for i, c := range cargos {
		result[i] = assemble(c, ctx.handlingEventRepository)
	}

	b, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func cargoHandler(w http.ResponseWriter, req *http.Request, ctx *context) {
	if req.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(req)

	c, err := ctx.cargoRepository.Find(cargo.TrackingID(vars["id"]))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	b, err := json.Marshal(assemble(c, ctx.handlingEventRepository))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func requestRoutesHandler(w http.ResponseWriter, req *http.Request, ctx *context) {
	if req.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(req)

	itineraries := ctx.bookingService.RequestPossibleRoutesForCargo(cargo.TrackingID(vars["id"]))

	var result []routeView
	for _, itin := range itineraries {
		var legs []legView
		for _, leg := range itin.Legs {
			legs = append(legs, legView{
				VoyageNumber: string(leg.VoyageNumber),
				From:         string(leg.LoadLocation),
				To:           string(leg.UnloadLocation),
				LoadTime:     leg.LoadTime,
				UnloadTime:   leg.UnloadTime,
			})
		}
		result = append(result, routeView{Legs: legs})
	}

	b, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func assignToRouteHandler(w http.ResponseWriter, req *http.Request, ctx *context) {
	if req.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	var route routeView
	if err := json.Unmarshal(b, &route); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var legs []cargo.Leg
	for _, l := range route.Legs {
		legs = append(legs, cargo.Leg{
			VoyageNumber:   voyage.Number(l.VoyageNumber),
			LoadLocation:   location.UNLocode(l.From),
			UnloadLocation: location.UNLocode(l.To),
			LoadTime:       l.LoadTime,
			UnloadTime:     l.UnloadTime,
		})
	}

	vars := mux.Vars(req)

	if err := ctx.bookingService.AssignCargoToRoute(cargo.Itinerary{Legs: legs}, cargo.TrackingID(vars["id"])); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func changeDestinationHandler(w http.ResponseWriter, req *http.Request, ctx *context) {
	if req.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	v := queryParams(req.URL.Query())

	found, missing := v.validateQueryParams("destination")

	if len(missing) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	vars := mux.Vars(req)

	if err := ctx.bookingService.ChangeDestination(cargo.TrackingID(vars["id"]), location.UNLocode(found["destination"].(string))); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func bookCargoHandler(w http.ResponseWriter, req *http.Request, ctx *context) {
	if req.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	v := queryParams(req.URL.Query())

	found, missing := v.validateQueryParams("origin", "destination", "arrivalDeadline")

	if len(missing) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var (
		origin          = found["origin"].(string)
		destination     = found["destination"].(string)
		arrivalDeadline = found["arrivalDeadline"].(string)
	)

	millis, _ := strconv.ParseInt(arrivalDeadline, 10, 64)
	trackingID, err := ctx.bookingService.BookNewCargo(location.UNLocode(origin), location.UNLocode(destination), time.Unix(millis/1000, 0))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c, err := ctx.cargoRepository.Find(trackingID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	b, err := json.Marshal(assemble(c, ctx.handlingEventRepository))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func locationsHandler(w http.ResponseWriter, req *http.Request, ctx *context) {
	if req.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	locations := ctx.locationRepository.FindAll()

	result := make([]locationView, len(locations))
	for i, loc := range locations {
		result[i] = locationView{
			UNLocode: string(loc.UNLocode),
			Name:     loc.Name,
		}
	}

	b, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func registerIncidentHandler(w http.ResponseWriter, req *http.Request, ctx *context) {
	if req.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	v := queryParams(req.URL.Query())

	found, missing := v.validateQueryParams("completionTime", "trackingId", "voyage", "location", "eventType")

	if len(missing) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var (
		completionTime = found["completionTime"].(string)
		trackingID     = found["trackingId"].(string)
		voyageNumber   = found["voyage"].(string)
		unLocode       = found["location"].(string)
		eventType      = found["eventType"].(string)
	)

	millis, _ := strconv.ParseInt(completionTime, 10, 64)
	if err := ctx.handlingEventService.RegisterHandlingEvent(time.Unix(millis/1000, 0), cargo.TrackingID(trackingID), voyage.Number(voyageNumber), location.UNLocode(unLocode), stringToEventType(eventType)); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func stringToEventType(s string) cargo.HandlingEventType {
	types := map[string]cargo.HandlingEventType{
		cargo.Receive.String(): cargo.Receive,
		cargo.Load.String():    cargo.Load,
		cargo.Unload.String():  cargo.Unload,
		cargo.Customs.String(): cargo.Customs,
		cargo.Claim.String():   cargo.Claim,
	}
	return types[s]
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
