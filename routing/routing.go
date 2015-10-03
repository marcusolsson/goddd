package routing

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/voyage"
)

// Service provides access to an external routing service.
type Service interface {
	// FetchRoutesForSpecification finds all possible routes that satisfy a
	// given specification.
	FetchRoutesForSpecification(routeSpecification cargo.RouteSpecification) []cargo.Itinerary
}

type routeResponse struct {
	Paths []routeDTO `json:"paths"`
}

type routeDTO struct {
	Edges []edgeDTO `json:"edges"`
}

type edgeDTO struct {
	Voyage      string    `json:"voyage"`
	Origin      string    `json:"origin"`
	Destination string    `json:"destination"`
	Departure   time.Time `json:"departure"`
	Arrival     time.Time `json:"arrival"`
}

type service struct {
	addr string
}

func (s *service) FetchRoutesForSpecification(routeSpecification cargo.RouteSpecification) []cargo.Itinerary {
	query := "from=" + string(routeSpecification.Origin) + "&to=" + string(routeSpecification.Destination)
	resp, err := http.Get(s.addr + "/paths?" + query)

	if err != nil {
		return []cargo.Itinerary{}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	var routes routeResponse
	err = json.Unmarshal(body, &routes)

	if err != nil {
		return []cargo.Itinerary{}
	}

	var itineraries []cargo.Itinerary
	for _, r := range routes.Paths {
		var legs []cargo.Leg
		for _, e := range r.Edges {
			legs = append(legs, cargo.Leg{
				VoyageNumber:   voyage.Number(e.Voyage),
				LoadLocation:   location.UNLocode(e.Origin),
				UnloadLocation: location.UNLocode(e.Destination),
				LoadTime:       e.Departure,
				UnloadTime:     e.Arrival,
			})
		}

		itineraries = append(itineraries, cargo.Itinerary{Legs: legs})
	}

	return itineraries
}

// NewService returns a new instance of the routing service.
func NewService(a string) Service {
	return &service{addr: a}
}
