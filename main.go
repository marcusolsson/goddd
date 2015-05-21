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
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/repository"
	"github.com/marcusolsson/goddd/routing"
	"github.com/marcusolsson/goddd/voyage"
)

var (
	cargoRepository            = repository.NewCargo()
	locationRepository         = repository.NewLocation()
	voyageRepository           = repository.NewVoyage()
	handlingEventRepository    = repository.NewHandlingEvent()
	routingService             = routing.NewService("http://ddd-pathfinder.herokuapp.com")
	bookingService             = booking.NewService(cargoRepository, locationRepository, routingService)
	handlingEventServiceFacade = handling.NewFacade(cargoRepository, locationRepository, voyageRepository, handlingEventRepository)
)

var (
	defaultPort = "8080"
)

func main() {
	storeTestData(cargoRepository)

	r := mux.NewRouter()

	r.HandleFunc("/cargos", cargosHandler).Methods("GET")
	r.HandleFunc("/cargos", bookCargoHandler).Methods("POST")
	r.HandleFunc("/cargos/{id}", cargoHandler).Methods("GET")
	r.HandleFunc("/cargos/{id}/request_routes", requestRoutesHandler).Methods("GET")
	r.HandleFunc("/cargos/{id}/assign_to_route", assignToRouteHandler).Methods("POST")
	r.HandleFunc("/cargos/{id}/change_destination", changeDestinationHandler).Methods("POST")
	r.HandleFunc("/locations", locationsHandler).Methods("GET")
	r.HandleFunc("/incidents", registerIncidentHandler).Methods("POST")

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

func cargosHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	cargos := cargoRepository.FindAll()

	dtos := make([]booking.CargoDTO, len(cargos))
	for i, c := range cargos {
		dtos[i] = booking.Assemble(c, handlingEventRepository)
	}

	b, err := json.Marshal(dtos)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func cargoHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(req)

	c, err := cargoRepository.Find(cargo.TrackingID(vars["id"]))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	b, err := json.Marshal(booking.Assemble(c, handlingEventRepository))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func requestRoutesHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(req)

	itineraries := bookingService.RequestPossibleRoutesForCargo(cargo.TrackingID(vars["id"]))

	var candidates []booking.RouteCandidateDTO
	for _, itin := range itineraries {
		var legs []booking.LegDTO
		for _, leg := range itin.Legs {
			legs = append(legs, booking.LegDTO{
				VoyageNumber: string(leg.VoyageNumber),
				From:         string(leg.LoadLocation),
				To:           string(leg.UnloadLocation),
				LoadTime:     leg.LoadTime,
				UnloadTime:   leg.UnloadTime,
			})
		}
		candidates = append(candidates, booking.RouteCandidateDTO{Legs: legs})
	}

	b, err := json.Marshal(candidates)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func assignToRouteHandler(w http.ResponseWriter, req *http.Request) {
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

	var candidate booking.RouteCandidateDTO
	if err := json.Unmarshal(b, &candidate); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var legs []cargo.Leg
	for _, l := range candidate.Legs {
		legs = append(legs, cargo.Leg{
			VoyageNumber:   voyage.Number(l.VoyageNumber),
			LoadLocation:   location.UNLocode(l.From),
			UnloadLocation: location.UNLocode(l.To),
			LoadTime:       l.LoadTime,
			UnloadTime:     l.UnloadTime,
		})
	}

	vars := mux.Vars(req)

	if err := bookingService.AssignCargoToRoute(cargo.Itinerary{Legs: legs}, cargo.TrackingID(vars["id"])); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func changeDestinationHandler(w http.ResponseWriter, req *http.Request) {
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

	if err := bookingService.ChangeDestination(cargo.TrackingID(vars["id"]), location.UNLocode(found["destination"].(string))); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func bookCargoHandler(w http.ResponseWriter, req *http.Request) {
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
	trackingID, err := bookingService.BookNewCargo(location.UNLocode(origin), location.UNLocode(destination), time.Unix(millis/1000, 0))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c, err := cargoRepository.Find(trackingID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	dto := booking.Assemble(c, handlingEventRepository)

	b, err := json.Marshal(dto)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func locationsHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	locations := locationRepository.FindAll()

	dtos := make([]booking.LocationDTO, len(locations))
	for i, loc := range locations {
		dtos[i] = booking.LocationDTO{
			UNLocode: string(loc.UNLocode),
			Name:     loc.Name,
		}
	}

	b, err := json.Marshal(dtos)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func registerIncidentHandler(w http.ResponseWriter, req *http.Request) {
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
		location       = found["location"].(string)
		eventType      = found["eventType"].(string)
	)

	if err := handlingEventServiceFacade.RegisterHandlingEvent(completionTime, trackingID, voyageNumber, location, eventType); err != nil {
		w.WriteHeader(http.StatusBadRequest)
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
