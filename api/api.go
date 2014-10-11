package api

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/marcusolsson/goddd/application"
	"github.com/marcusolsson/goddd/domain/cargo"
	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/infrastructure"
	"github.com/marcusolsson/goddd/interfaces"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/cors"
	"github.com/martini-contrib/render"
)

type jsonObject map[string]interface{}

func RegisterHandlers() {
	var (
		cargoRepository      = infrastructure.NewInMemCargoRepository()
		locationRepository   = infrastructure.NewInMemLocationRepository()
		routingService       = infrastructure.NewExternalRoutingService(locationRepository)
		bookingService       = application.NewBookingService(cargoRepository, locationRepository, routingService)
		bookingServiceFacade = interfaces.NewBookingServiceFacade(cargoRepository, locationRepository, bookingService)
	)

	// Store some sample cargos.
	storeTestData(cargoRepository)

	var (
		ResourceNotFound              = jsonObject{"error": "The specified resource does not exist."}
		MissingRequiredQueryParameter = jsonObject{"error": "A required query parameter was not specified for this request."}
		InvalidInput                  = jsonObject{"error": "One of the request inputs is not valid."}
	)

	m := martini.Classic()

	m.Use(martini.Static("app"))

	m.Use(cors.Allow(&cors.Options{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type"},
	}))

	m.Use(render.Renderer(render.Options{
		IndentJSON: true,
	}))

	// GET /cargos
	m.Get("/cargos", func(r render.Render) {
		r.JSON(200, bookingServiceFacade.ListAllCargos())
	})

	m.Get("/cargos/:id", func(params martini.Params, r render.Render) {
		c, err := bookingServiceFacade.LoadCargoForRouting(params["id"])

		if err != nil {
			r.JSON(404, ResourceNotFound)
			return
		}

		r.JSON(200, c)
	})

	// POST /cargos/:id/change_destination
	m.Post("/cargos/:id/change_destination", func(req *http.Request, params martini.Params, r render.Render) {
		v := queryParams(req.URL.Query())
		found, missing := v.validateQueryParams("destination")

		if len(missing) > 0 {
			e := MissingRequiredQueryParameter
			e["missing"] = missing
			r.JSON(400, e)
			return
		}

		if err := bookingServiceFacade.ChangeDestination(params["id"], fmt.Sprintf("%s", found["destination"])); err != nil {
			r.JSON(400, InvalidInput)
			return
		}

		r.JSON(200, jsonObject{})
	})

	// POST /cargos/:id/assign_to_route
	m.Post("/cargos/:id/assign_to_route", binding.Bind(interfaces.RouteCandidateDTO{}), func(c interfaces.RouteCandidateDTO, params martini.Params, r render.Render) {
		if err := bookingServiceFacade.AssignCargoToRoute(params["id"], c); err != nil {
			r.JSON(400, InvalidInput)
			return
		}

		r.JSON(200, jsonObject{})
	})

	// GET /cargos/:id/request_routes
	m.Get("/cargos/:id/request_routes", func(params martini.Params, r render.Render) {
		r.JSON(200, bookingServiceFacade.RequestRoutesForCargo(params["id"]))
	})

	// POST /cargos
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
			origin          = found["origin"].(string)
			destination     = found["destination"].(string)
			arrivalDeadline = found["arrivalDeadline"].(string)
		)

		trackingId, err := bookingServiceFacade.BookNewCargo(origin, destination, arrivalDeadline)

		if err != nil {
			r.JSON(400, InvalidInput)
			return
		}

		c, err := bookingServiceFacade.LoadCargoForRouting(trackingId)

		if err != nil {
			r.JSON(404, ResourceNotFound)
			return
		}

		r.JSON(200, c)
	})

	// GET /locations
	m.Get("/locations", func(r render.Render) {
		r.JSON(200, bookingServiceFacade.ListShippingLocations())
	})

	http.Handle("/", m)
}

func storeTestData(r cargo.CargoRepository) {
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
