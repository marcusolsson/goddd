package booking

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/voyage"
)

type cargoDTO struct {
	TrackingID           string     `json:"trackingId"`
	StatusText           string     `json:"statusText"`
	Origin               string     `json:"origin"`
	Destination          string     `json:"destination"`
	ETA                  time.Time  `json:"eta"`
	NextExpectedActivity string     `json:"nextExpectedActivity"`
	Misrouted            bool       `json:"misrouted"`
	Routed               bool       `json:"routed"`
	ArrivalDeadline      time.Time  `json:"arrivalDeadline"`
	Events               []eventDTO `json:"events"`
	Legs                 []legDTO   `json:"legs,omitempty"`
}

type locationDTO struct {
	UNLocode string `json:"locode"`
	Name     string `json:"name"`
}

type RouteCandidateDTO struct {
	Legs []legDTO `json:"legs"`
}

type legDTO struct {
	VoyageNumber string    `json:"voyageNumber"`
	From         string    `json:"from"`
	To           string    `json:"to"`
	LoadTime     time.Time `json:"loadTime"`
	UnloadTime   time.Time `json:"unloadTime"`
}

type eventDTO struct {
	Description string `json:"description"`
	Expected    bool   `json:"expected"`
}

type Facade interface {
	BookNewCargo(origin, destination string, arrivalDeadline string) (string, error)
	LoadCargoForRouting(trackingID string) (cargoDTO, error)
	AssignCargoToRoute(trackingID string, candidate RouteCandidateDTO) error
	ChangeDestination(trackingID string, destinationUNLocode string) error
	RequestRoutesForCargo(trackingID string) []RouteCandidateDTO
	ListShippingLocations() []locationDTO
	ListAllCargos() []cargoDTO
}

type facade struct {
	cargoRepository         cargo.Repository
	locationRepository      location.Repository
	handlingEventRepository cargo.HandlingEventRepository
	Service                 Service
}

func (f *facade) BookNewCargo(origin, destination string, arrivalDeadline string) (string, error) {
	millis, _ := strconv.ParseInt(fmt.Sprintf("%s", arrivalDeadline), 10, 64)
	trackingID, err := f.Service.BookNewCargo(location.UNLocode(origin), location.UNLocode(destination), time.Unix(millis/1000, 0))

	return string(trackingID), err
}

func (f *facade) LoadCargoForRouting(trackingID string) (cargoDTO, error) {
	c, err := f.cargoRepository.Find(cargo.TrackingID(trackingID))

	if err != nil {
		return cargoDTO{}, err
	}

	return assemble(c, f.handlingEventRepository), nil
}

func (f *facade) AssignCargoToRoute(trackingID string, candidate RouteCandidateDTO) error {
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

	return f.Service.AssignCargoToRoute(cargo.Itinerary{Legs: legs}, cargo.TrackingID(trackingID))
}

func (f *facade) ChangeDestination(trackingID string, destinationUNLocode string) error {
	return f.Service.ChangeDestination(cargo.TrackingID(trackingID), location.UNLocode(destinationUNLocode))
}

func (f *facade) RequestRoutesForCargo(trackingID string) []RouteCandidateDTO {
	itineraries := f.Service.RequestPossibleRoutesForCargo(cargo.TrackingID(trackingID))

	var candidates []RouteCandidateDTO
	for _, itin := range itineraries {
		var legs []legDTO
		for _, leg := range itin.Legs {
			legs = append(legs, legDTO{
				VoyageNumber: string(leg.VoyageNumber),
				From:         string(leg.LoadLocation),
				To:           string(leg.UnloadLocation),
				LoadTime:     leg.LoadTime,
				UnloadTime:   leg.UnloadTime,
			})
		}
		candidates = append(candidates, RouteCandidateDTO{Legs: legs})
	}

	return candidates
}

func (f *facade) ListShippingLocations() []locationDTO {
	locations := f.locationRepository.FindAll()

	dtos := make([]locationDTO, len(locations))
	for i, loc := range locations {
		dtos[i] = locationDTO{
			UNLocode: string(loc.UNLocode),
			Name:     loc.Name,
		}
	}

	return dtos
}

func (f *facade) ListAllCargos() []cargoDTO {
	cargos := f.cargoRepository.FindAll()
	dtos := make([]cargoDTO, len(cargos))

	for i, c := range cargos {
		dtos[i] = assemble(c, f.handlingEventRepository)
	}

	return dtos
}

func NewFacade(s Service) Facade {
	return &facade{Service: s}
}

func assemble(c cargo.Cargo, her cargo.HandlingEventRepository) cargoDTO {
	return cargoDTO{
		TrackingID:           string(c.TrackingID),
		Origin:               string(c.Origin),
		Destination:          string(c.RouteSpecification.Destination),
		ETA:                  c.Delivery.ETA,
		NextExpectedActivity: nextExpectedActivity(c),
		Misrouted:            c.Delivery.RoutingStatus == cargo.Misrouted,
		Routed:               !c.Itinerary.IsEmpty(),
		ArrivalDeadline:      c.RouteSpecification.ArrivalDeadline,
		StatusText:           assembleStatusText(c),
		Legs:                 assembleLegs(c),
		Events:               assembleEvents(c, her),
	}
}

func nextExpectedActivity(c cargo.Cargo) string {
	a := c.Delivery.NextExpectedActivity
	prefix := "Next expected activity is to"

	switch a.Type {
	case cargo.Load:
		return fmt.Sprintf("%s %s cargo onto voyage %s in %s.", prefix, strings.ToLower(a.Type.String()), a.VoyageNumber, a.Location)
	case cargo.Unload:
		return fmt.Sprintf("%s %s cargo off of voyage %s in %s.", prefix, strings.ToLower(a.Type.String()), a.VoyageNumber, a.Location)
	case cargo.NotHandled:
		return "There are currently no expected activities for this cargo."
	}

	return fmt.Sprintf("%s %s cargo in %s.", prefix, strings.ToLower(a.Type.String()), a.Location)
}

func assembleStatusText(c cargo.Cargo) string {
	switch c.Delivery.TransportStatus {
	case cargo.NotReceived:
		return "Not received"
	case cargo.InPort:
		return fmt.Sprintf("In port %s", c.Delivery.LastKnownLocation)
	case cargo.OnboardCarrier:
		return fmt.Sprintf("Onboard voyage %s", c.Delivery.CurrentVoyage)
	case cargo.Claimed:
		return "Claimed"
	default:
		return "Unknown"
	}
}

func assembleLegs(c cargo.Cargo) []legDTO {
	var legs []legDTO
	for _, l := range c.Itinerary.Legs {
		legs = append(legs, legDTO{
			VoyageNumber: string(l.VoyageNumber),
			From:         string(l.LoadLocation),
			To:           string(l.UnloadLocation),
			LoadTime:     l.LoadTime,
			UnloadTime:   l.UnloadTime,
		})
	}
	return legs
}

func assembleEvents(c cargo.Cargo, r cargo.HandlingEventRepository) []eventDTO {
	h := r.QueryHandlingHistory(c.TrackingID)
	events := make([]eventDTO, len(h.HandlingEvents))
	for i, e := range h.HandlingEvents {
		var description string

		switch e.Activity.Type {
		case cargo.NotHandled:
			description = "Cargo has not yet been received."
		case cargo.Receive:
			description = fmt.Sprintf("Received in %s, at %s", e.Activity.Location, time.Now().Format(time.RFC3339))
		case cargo.Load:
			description = fmt.Sprintf("Loaded onto voyage %s in %s, at %s.", e.Activity.VoyageNumber, e.Activity.Location, time.Now().Format(time.RFC3339))
		case cargo.Unload:
			description = fmt.Sprintf("Unloaded off voyage %s in %s, at %s.", e.Activity.VoyageNumber, e.Activity.Location, time.Now().Format(time.RFC3339))
		case cargo.Claim:
			description = fmt.Sprintf("Claimed in %s, at %s.", e.Activity.Location, time.Now().Format(time.RFC3339))
		case cargo.Customs:
			description = fmt.Sprintf("Cleared customs in %s, at %s.", e.Activity.Location, time.Now().Format(time.RFC3339))
		default:
			description = "[Unknown status]"
		}

		events[i] = eventDTO{Description: description, Expected: c.Itinerary.IsExpected(e)}
	}

	return events
}
