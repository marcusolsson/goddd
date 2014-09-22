package server

import (
	"fmt"
	"net/http"

	"bitbucket.org/marcus_olsson/goddd/cargo"
	"bitbucket.org/marcus_olsson/goddd/location"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

type Event struct {
	Description string `json:"description"`
	Expected    bool   `json:"expected"`
}

type cargoDTO struct {
	TrackingId           string  `json:"trackingId"`
	StatusText           string  `json:"statusText"`
	Destination          string  `json:"destination"`
	ETA                  string  `json:"eta"`
	NextExpectedActivity string  `json:"nextExpectedActivity"`
	Events               []Event `json:"events"`
}

func Assemble(c cargo.Cargo) cargoDTO {
	dto := cargoDTO{
		TrackingId:           string(c.TrackingId),
		StatusText:           fmt.Sprintf("%s %s", cargo.InPort, c.Origin.Name),
		Destination:          c.RouteSpecification.Destination.Name,
		ETA:                  "2009-03-12 12:00",
		NextExpectedActivity: "Next expected activity is to load cargo onto voyage 0200T in New York",
	}

	dto.Events = make([]Event, 3)
	dto.Events[0] = Event{Description: "Received in Hongkong, at 3/1/09 12:00 AM.", Expected: true}
	dto.Events[1] = Event{Description: "Loaded onto voyage 0100S in Hongkong, at 3/2/09 12:00 AM."}
	dto.Events[2] = Event{Description: "Unloaded off voyage 0100S in New York, at 3/5/09 12:00 AM."}

	return dto
}

type JSONObject map[string]interface{}

var (
	ResourceNotFound = JSONObject{"error": "resource not found"}
)

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

func RegisterHandlers() {
	repository := cargo.NewCargoRepository()
	storeTestData(repository)

	m := martini.Classic()

	m.Use(martini.Static("app"))
	m.Use(render.Renderer(render.Options{
		IndentJSON: true,
	}))

	m.Get("/cargos", func(r render.Render) {
		cargos := repository.FindAll()
		dtos := make([]cargoDTO, len(cargos))
		for i, c := range cargos {
			dtos[i] = Assemble(c)
		}
		r.JSON(200, dtos)
	})

	m.Get("/cargos/:id", func(params martini.Params, r render.Render) {
		c, err := repository.Find(cargo.TrackingId(params["id"]))

		if err != nil {
			r.JSON(404, ResourceNotFound)
		} else {
			r.JSON(200, Assemble(*c))
		}

	})

	http.Handle("/", m)
}
