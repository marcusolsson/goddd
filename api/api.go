package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/marcusolsson/goddd/application/booking"
	"github.com/marcusolsson/goddd/domain/cargo"
	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/routing"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/cors"
	"github.com/martini-contrib/render"
)

type eventDTO struct {
	Description string `json:"description"`
	Expected    bool   `json:"expected"`
}

type locationDTO struct {
	UNLocode string `json:"locode"`
	Name     string `json:"name"`
}

type cargoDTO struct {
	TrackingId           string     `json:"trackingId"`
	StatusText           string     `json:"statusText"`
	Origin               string     `json:"origin"`
	Destination          string     `json:"destination"`
	ETA                  string     `json:"eta"`
	NextExpectedActivity string     `json:"nextExpectedActivity"`
	Misrouted            bool       `json:"misrouted"`
	Routed               bool       `json:"routed"`
	ArrivalDeadline      string     `json:"arrivalDeadline"`
	Events               []eventDTO `json:"events"`
	Legs                 []legDTO   `json:"legs"`
}

type legDTO struct {
	VoyageNumber string `json:"voyage"`
	From         string `json:"from"`
	To           string `json:"to"`
	LoadTime     string `json:"loadTime"`
	UnloadTime   string `json:"unloadTime"`
}

type routeCandidate struct {
	Legs []legDTO `json:"legs"`
}

func assemble(c cargo.Cargo) cargoDTO {
	eta := time.Date(2009, time.March, 12, 12, 0, 0, 0, time.UTC)
	dto := cargoDTO{
		TrackingId:           string(c.TrackingId),
		StatusText:           fmt.Sprintf("%s %s", cargo.InPort, c.Origin.Name),
		Origin:               string(c.Origin.UNLocode),
		Destination:          string(c.RouteSpecification.Destination.UNLocode),
		ETA:                  eta.Format(time.RFC3339),
		NextExpectedActivity: "Next expected activity is to load cargo onto voyage 0200T in New York",
		Misrouted:            c.Delivery.RoutingStatus == cargo.Misrouted,
		Routed:               !c.Itinerary.IsEmpty(),
		ArrivalDeadline:      c.ArrivalDeadline.Format(time.RFC3339),
	}

	legs := make([]legDTO, 0)
	for _, l := range c.Itinerary.Legs {
		legs = append(legs, legDTO{
			VoyageNumber: l.Voyage,
			From:         string(l.LoadLocation.UNLocode),
			To:           string(l.UnloadLocation.UNLocode),
		})
	}
	dto.Legs = legs

	dto.Events = make([]eventDTO, 3)
	dto.Events[0] = eventDTO{Description: "Received in Hongkong, at 3/1/09 12:00 AM.", Expected: true}
	dto.Events[1] = eventDTO{Description: "Loaded onto voyage 0100S in Hongkong, at 3/2/09 12:00 AM."}
	dto.Events[2] = eventDTO{Description: "Unloaded off voyage 0100S in New York, at 3/5/09 12:00 AM."}

	return dto
}

type jsonObject map[string]interface{}

func RegisterHandlers() {
	var (
		cargoRepository    = cargo.NewCargoRepository()
		locationRepository = location.NewLocationRepository()
		routingService     = routing.NewRoutingService(locationRepository)
		bookingService     = booking.NewBookingService(cargoRepository, locationRepository, routingService)
	)

	var (
		ResourceNotFound              = jsonObject{"error": "The specified resource does not exist."}
		MissingRequiredQueryParameter = jsonObject{"error": "A required query parameter was not specified for this request."}
		InvalidInput                  = jsonObject{"error": "One of the request inputs is not valid."}
	)

	// Store some sample cargos.
	storeTestData(cargoRepository)

	m := martini.Classic()

	m.Use(cors.Allow(&cors.Options{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type"},
	}))

	m.Use(martini.Static("app"))
	m.Use(render.Renderer(render.Options{
		IndentJSON: true,
	}))

	m.Get("/cargos", func(r render.Render) {
		cargos := cargoRepository.FindAll()
		dtos := make([]cargoDTO, len(cargos))

		for i, c := range cargos {
			dtos[i] = assemble(c)
		}

		r.JSON(200, dtos)
	})

	m.Get("/cargos/:id", func(params martini.Params, r render.Render) {
		trackingId := cargo.TrackingId(params["id"])
		c, err := cargoRepository.Find(trackingId)

		if err != nil {
			r.JSON(404, ResourceNotFound)
		} else {
			r.JSON(200, assemble(c))
		}
	})

	m.Post("/cargos/:id/change_destination", func(req *http.Request, params martini.Params, r render.Render) {
		v := queryParams(req.URL.Query())
		found, missing := v.validateQueryParams("destination")

		if len(missing) > 0 {
			e := MissingRequiredQueryParameter
			e["missing"] = missing
			r.JSON(400, e)
			return
		}

		var (
			trackingId  = cargo.TrackingId(params["id"])
			destination = location.UNLocode(fmt.Sprintf("%s", found["destination"]))
		)

		if err := bookingService.ChangeDestination(trackingId, destination); err != nil {
			r.JSON(400, InvalidInput)
			return
		}

		r.JSON(200, jsonObject{})
	})

	m.Post("/cargos/:id/assign_to_route", binding.Bind(routeCandidate{}), func(rc routeCandidate, params martini.Params, r render.Render) {
		trackingId := cargo.TrackingId(params["id"])

		legs := make([]cargo.Leg, 0)
		for _, l := range rc.Legs {

			var (
				loadLocation   = locationRepository.Find(location.UNLocode(l.From))
				unloadLocation = locationRepository.Find(location.UNLocode(l.To))
			)

			legs = append(legs, cargo.Leg{
				Voyage:         l.VoyageNumber,
				LoadLocation:   loadLocation,
				UnloadLocation: unloadLocation,
			})
		}

		itinerary := cargo.Itinerary{Legs: legs}

		if err := bookingService.AssignCargoToRoute(itinerary, trackingId); err != nil {
			r.JSON(400, InvalidInput)
			return
		}

		r.JSON(200, itinerary)
	})

	m.Get("/cargos/:id/request_routes", func(params martini.Params, r render.Render) {
		trackingId := cargo.TrackingId(params["id"])
		itineraries := bookingService.RequestPossibleRoutesForCargo(trackingId)

		candidates := make([]routeCandidate, 0)
		for _, itin := range itineraries {
			legs := make([]legDTO, 0)
			for _, leg := range itin.Legs {
				legs = append(legs, legDTO{
					VoyageNumber: "S0001",
					From:         string(leg.LoadLocation.UNLocode),
					To:           string(leg.UnloadLocation.UNLocode),
					LoadTime:     "N/A",
					UnloadTime:   "N/A",
				})
			}
			candidates = append(candidates, routeCandidate{Legs: legs})
		}

		r.JSON(200, candidates)
	})

	m.Post("/cargos", func(req *http.Request, r render.Render) {
		v := queryParams(req.URL.Query())
		found, missing := v.validateQueryParams("origin", "destination", "arrivalDeadline")

		if len(missing) > 0 {
			e := MissingRequiredQueryParameter
			e["missing"] = missing
			r.JSON(400, e)
			return
		}

		var (
			origin      = location.UNLocode(fmt.Sprintf("%s", found["origin"]))
			destination = location.UNLocode(fmt.Sprintf("%s", found["destination"]))
		)

		millis, _ := strconv.ParseInt(fmt.Sprintf("%s", found["arrivalDeadline"]), 10, 64)
		arrivalDeadline := time.Unix(millis/1000, 0)

		trackingId, err := bookingService.BookNewCargo(origin, destination, arrivalDeadline)

		if err != nil {
			r.JSON(400, InvalidInput)
			return
		}

		c, err := cargoRepository.Find(trackingId)

		if err != nil {
			r.JSON(404, ResourceNotFound)
			return
		}

		r.JSON(200, assemble(c))
	})

	m.Get("/locations", func(r render.Render) {
		locationRepository := location.NewLocationRepository()
		locations := locationRepository.FindAll()

		dtos := make([]locationDTO, len(locations))
		for i, loc := range locations {
			dtos[i] = locationDTO{
				UNLocode: string(loc.UNLocode),
				Name:     loc.Name,
			}
		}

		r.JSON(200, dtos)
	})

	http.Handle("/", m)
}

func storeTestData(r cargo.CargoRepository) {
	test1 := cargo.NewCargo("FTL456", cargo.RouteSpecification{
		Origin:      location.Melbourne,
		Destination: location.Stockholm,
	})
	r.Store(*test1)

	test2 := cargo.NewCargo("ABC123", cargo.RouteSpecification{
		Origin:      location.Stockholm,
		Destination: location.Hongkong,
	})
	r.Store(*test2)
}

type queryParams url.Values

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
