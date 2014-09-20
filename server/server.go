package main

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

type Event struct {
	Description string `json:"description"`
	Expected    bool   `json:"expected"`
}

type Cargo struct {
	TrackingId           string  `json:"trackingId"`
	StatusText           string  `json:"statusText"`
	Destination          string  `json:"destination"`
	ETA                  string  `json:"eta"`
	NextExpectedActivity string  `json:"nextExpectedActivity"`
	Events               []Event `json:"events"`
}

func main() {
	m := martini.Classic()

	m.Use(render.Renderer(render.Options{
		IndentJSON: true,
	}))

	m.Get("/cargos/:id", func(params martini.Params, r render.Render) {
		c := Cargo{
			TrackingId:           params["id"],
			StatusText:           "In port New York",
			Destination:          "Helsinki",
			ETA:                  "2009-03-12 12:00",
			NextExpectedActivity: "Next expected activity is to load cargo onto voyage 0200T in New York",
		}

		c.Events = make([]Event, 3)
		c.Events[0] = Event{Description: "Received in Hongkong, at 3/1/09 12:00 AM.", Expected: true}
		c.Events[1] = Event{Description: "Loaded onto voyage 0100S in Hongkong, at 3/2/09 12:00 AM."}
		c.Events[2] = Event{Description: "Unloaded off voyage 0100S in New York, at 3/5/09 12:00 AM."}

		r.JSON(200, c)
	})

	m.Run()
}
