package tracking

import (
	"fmt"
	"strings"
	"time"

	"github.com/marcusolsson/goddd/cargo"
)

type Service interface {
	Track(id string) (Cargo, error)
}

type service struct {
	cargos         cargo.Repository
	handlingEvents cargo.HandlingEventRepository
}

func (s *service) Track(id string) (Cargo, error) {
	c, err := s.cargos.Find(cargo.TrackingID(id))
	if err != nil {
		return Cargo{}, err
	}
	return assemble(c, s.handlingEvents), nil
}

func NewService() Service {
	return &service{}
}

type Cargo struct {
	TrackingID           string      `json:"trackingId"`
	StatusText           string      `json:"statusText"`
	Origin               string      `json:"origin"`
	Destination          string      `json:"destination"`
	ETA                  time.Time   `json:"eta"`
	NextExpectedActivity string      `json:"nextExpectedActivity"`
	Misrouted            bool        `json:"misrouted"`
	Routed               bool        `json:"routed"`
	ArrivalDeadline      time.Time   `json:"arrivalDeadline"`
	Events               []eventView `json:"events"`
	Legs                 []legView   `json:"legs,omitempty"`
}

type routeView struct {
	Legs []legView `json:"legs"`
}

type legView struct {
	VoyageNumber string    `json:"voyageNumber"`
	From         string    `json:"from"`
	To           string    `json:"to"`
	LoadTime     time.Time `json:"loadTime"`
	UnloadTime   time.Time `json:"unloadTime"`
}

type eventView struct {
	Description string `json:"description"`
	Expected    bool   `json:"expected"`
}

func assemble(c cargo.Cargo, her cargo.HandlingEventRepository) Cargo {
	return Cargo{
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

func assembleLegs(c cargo.Cargo) []legView {
	var legs []legView
	for _, l := range c.Itinerary.Legs {
		legs = append(legs, legView{
			VoyageNumber: string(l.VoyageNumber),
			From:         string(l.LoadLocation),
			To:           string(l.UnloadLocation),
			LoadTime:     l.LoadTime,
			UnloadTime:   l.UnloadTime,
		})
	}
	return legs
}

func assembleEvents(c cargo.Cargo, r cargo.HandlingEventRepository) []eventView {
	h := r.QueryHandlingHistory(c.TrackingID)
	events := make([]eventView, len(h.HandlingEvents))
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

		events[i] = eventView{Description: description, Expected: c.Itinerary.IsExpected(e)}
	}

	return events
}
