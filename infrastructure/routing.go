package infrastructure

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/marcusolsson/goddd/domain/cargo"
	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/routing"
	"github.com/marcusolsson/goddd/domain/voyage"
)

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

type externalRoutingService struct {
	locationRepository location.Repository
}

func (s *externalRoutingService) FetchRoutesForSpecification(routeSpecification cargo.RouteSpecification) []cargo.Itinerary {

	query := "from=" + string(routeSpecification.Origin) + "&to=" + string(routeSpecification.Destination)
	resp, err := http.Get("http://ddd-pathfinder.herokuapp.com/paths?" + query)

	if err != nil {
		return []cargo.Itinerary{}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	var routes []routeDTO
	err = json.Unmarshal(body, &routes)

	if err != nil {
		return []cargo.Itinerary{}
	}

	var itineraries []cargo.Itinerary
	for _, r := range routes {
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

func NewExternalRoutingService(locationRepository location.Repository) routing.RoutingService {
	return &externalRoutingService{locationRepository: locationRepository}
}
