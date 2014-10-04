package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/marcusolsson/goddd/booking"
	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/routing"

	"github.com/go-martini/martini"
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
}

// Assemble converts the Cargo domain object to a serializable DTO.
func Assemble(c cargo.Cargo) cargoDTO {
	eta := time.Date(2009, time.March, 12, 12, 0, 0, 0, time.UTC)
	dto := cargoDTO{
		TrackingId:           string(c.TrackingId),
		StatusText:           fmt.Sprintf("%s %s", cargo.InPort, c.Origin.Name),
		Origin:               c.Origin.Name,
		Destination:          c.RouteSpecification.Destination.Name,
		ETA:                  eta.Format(time.RFC3339),
		NextExpectedActivity: "Next expected activity is to load cargo onto voyage 0200T in New York",
		Misrouted:            c.Delivery.RoutingStatus == cargo.Misrouted,
		Routed:               !c.Itinerary.IsEmpty(),
		ArrivalDeadline:      c.ArrivalDeadline.Format(time.RFC3339),
	}

	dto.Events = make([]eventDTO, 3)
	dto.Events[0] = eventDTO{Description: "Received in Hongkong, at 3/1/09 12:00 AM.", Expected: true}
	dto.Events[1] = eventDTO{Description: "Loaded onto voyage 0100S in Hongkong, at 3/2/09 12:00 AM."}
	dto.Events[2] = eventDTO{Description: "Unloaded off voyage 0100S in New York, at 3/5/09 12:00 AM."}

	return dto
}

type JSONObject map[string]interface{}

var (
	ResourceNotFound              = JSONObject{"error": "The specified resource does not exist."}
	MissingRequiredQueryParameter = JSONObject{"error": "A required query parameter was not specified for this request."}
	InvalidInput                  = JSONObject{"error": "One of the request inputs is not valid."}
)

// TODO: Globals are bad!
var (
	cargoRepository    = cargo.NewCargoRepository()
	locationRepository = location.NewLocationRepository()
	routingService     = routing.NewRoutingService()
	bookingService     = booking.NewBookingService(cargoRepository, locationRepository, routingService)
)

func RegisterHandlers() {

	// Store some sample cargos.
	storeTestData(cargoRepository)

	m := martini.Classic()

	allowCORSHandler := cors.Allow(&cors.Options{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST"},
		AllowHeaders: []string{"Origin"},
	})

	m.Use(martini.Static("app"))
	m.Use(render.Renderer(render.Options{
		IndentJSON: true,
	}))

	// GET /cargos
	// Returns an array of all booked cargos.
	m.Get("/cargos", allowCORSHandler, func(r render.Render) {
		cargos := cargoRepository.FindAll()
		dtos := make([]cargoDTO, len(cargos))

		for i, c := range cargos {
			dtos[i] = Assemble(c)
		}

		r.JSON(200, dtos)
	})

	// GET /cargos/:id
	// Finds and returns a cargo with a specified tracking id.
	m.Get("/cargos/:id", allowCORSHandler, func(params martini.Params, r render.Render) {
		trackingId := cargo.TrackingId(params["id"])
		c, err := cargoRepository.Find(trackingId)

		if err != nil {
			r.JSON(404, ResourceNotFound)
		} else {
			r.JSON(200, Assemble(c))
		}
	})

	// POST /cargos
	// Books a cargo from an origin to a destination within a specified arrival deadline.
	m.Post("/cargos", allowCORSHandler, func(req *http.Request, r render.Render) {
		v := QueryParams(req.URL.Query())
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

		r.JSON(200, Assemble(c))
	})

	// GET /locations
	// Returns an array of known locations.
	m.Get("/locations", allowCORSHandler, func(r render.Render) {
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

type QueryParams url.Values

func (p QueryParams) validateQueryParams(params ...string) (found JSONObject, missing []string) {
	found = make(JSONObject)
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
