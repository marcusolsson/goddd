package main

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

type Cargo struct {
	TrackingId           string `json:"trackingId"`
	StatusText           string `json:"statusText"`
	Destination          string `json:"destination"`
	ETA                  string `json:"eta"`
	NextExpectedActivity string `json:"nextExpectedActivity"`
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

		r.JSON(200, c)
	})

	m.Run()
}
