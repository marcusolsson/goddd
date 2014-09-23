package server

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

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

func NextTrackingId() cargo.TrackingId {
	f, _ := os.Open("/dev/urandom")
	b := make([]byte, 16)
	f.Read(b)
	f.Close()
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return cargo.TrackingId(strings.Split(strings.ToUpper(uuid), "-")[0])
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

type JSONObject map[string]interface{}

var (
	ResourceNotFound              = JSONObject{"error": "The specified resource does not exist."}
	MissingRequiredQueryParameter = JSONObject{"error": "A required query parameter was not specified for this request."}
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

	m.Post("/cargos", func(req *http.Request, r render.Render) {
		v := QueryParams(req.URL.Query())

		found, missing := v.validateQueryParams("origin", "destination", "arrivalDeadline")

		if len(missing) > 0 {
			e := MissingRequiredQueryParameter
			e["missing"] = missing
			r.JSON(400, e)
			return
		}

		lr := location.NewLocationRepository()

		orgn := lr.Find(location.UNLocode(fmt.Sprintf("%s", found["origin"])))

		if orgn == location.UnknownLocation {
			r.JSON(404, ResourceNotFound)
			return
		}

		dest := lr.Find(location.UNLocode(fmt.Sprintf("%s", found["destination"])))

		if dest == location.UnknownLocation {
			r.JSON(404, ResourceNotFound)
			return
		}

		c := cargo.NewCargo(NextTrackingId(), cargo.RouteSpecification{
			Origin:      orgn,
			Destination: dest,
		})

		repository.Store(*c)

		r.JSON(200, c)
	})

	http.Handle("/", m)
}
