package interfaces

import (
	"fmt"
	"strconv"
	"time"

	"github.com/marcusolsson/goddd/application"
	"github.com/marcusolsson/goddd/domain/cargo"
	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/voyage"
)

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
	Legs                 []legDTO   `json:"legs"`
}

type locationDTO struct {
	UNLocode string `json:"locode"`
	Name     string `json:"name"`
}

type RouteCandidateDTO struct {
	Legs []legDTO `json:"legs"`
}

type legDTO struct {
	VoyageNumber string `json:"voyageNumber"`
	From         string `json:"from"`
	To           string `json:"to"`
	LoadTime     string `json:"loadTime"`
	UnloadTime   string `json:"unloadTime"`
}

type eventDTO struct {
	Description string `json:"description"`
	Expected    bool   `json:"expected"`
}

type BookingServiceFacade interface {
	BookNewCargo(origin, destination string, arrivalDeadline string) (string, error)
	LoadCargoForRouting(trackingId string) (cargoDTO, error)
	AssignCargoToRoute(trackingId string, candidate RouteCandidateDTO) error
	ChangeDestination(trackingId string, destinationUNLocode string) error
	RequestRoutesForCargo(trackingId string) []RouteCandidateDTO
	ListShippingLocations() []locationDTO
	ListAllCargos() []cargoDTO
}

type bookingServiceFacade struct {
	cargoRepository    cargo.CargoRepository
	locationRepository location.LocationRepository
	bookingService     application.BookingService
}

func (f *bookingServiceFacade) BookNewCargo(origin, destination string, arrivalDeadline string) (string, error) {
	millis, _ := strconv.ParseInt(fmt.Sprintf("%s", arrivalDeadline), 10, 64)
	trackingId, err := f.bookingService.BookNewCargo(location.UNLocode(origin), location.UNLocode(destination), time.Unix(millis/1000, 0))

	return string(trackingId), err
}

func (f *bookingServiceFacade) LoadCargoForRouting(trackingId string) (cargoDTO, error) {
	c, err := f.cargoRepository.Find(cargo.TrackingId(trackingId))

	if err != nil {
		return cargoDTO{}, err
	}

	return assemble(c), nil
}

func (f *bookingServiceFacade) AssignCargoToRoute(trackingId string, candidate RouteCandidateDTO) error {
	legs := make([]cargo.Leg, 0)
	for _, l := range candidate.Legs {

		var (
			loadLocation   = f.locationRepository.Find(location.UNLocode(l.From))
			unloadLocation = f.locationRepository.Find(location.UNLocode(l.To))
			voyageNumber   = voyage.VoyageNumber(l.VoyageNumber)
		)

		legs = append(legs, cargo.Leg{
			VoyageNumber:   voyageNumber,
			LoadLocation:   loadLocation,
			UnloadLocation: unloadLocation,
		})
	}

	return f.bookingService.AssignCargoToRoute(cargo.Itinerary{legs}, cargo.TrackingId(trackingId))
}

func (f *bookingServiceFacade) ChangeDestination(trackingId string, destinationUNLocode string) error {
	return f.bookingService.ChangeDestination(cargo.TrackingId(trackingId), location.UNLocode(destinationUNLocode))
}

func (f *bookingServiceFacade) RequestRoutesForCargo(trackingId string) []RouteCandidateDTO {
	itineraries := f.bookingService.RequestPossibleRoutesForCargo(cargo.TrackingId(trackingId))

	candidates := make([]RouteCandidateDTO, 0)
	for _, itin := range itineraries {
		legs := make([]legDTO, 0)
		for _, leg := range itin.Legs {
			legs = append(legs, legDTO{
				VoyageNumber: "S0001",
				From:         string(leg.LoadLocation.UNLocode),
				To:           string(leg.UnloadLocation.UNLocode),
				LoadTime:     "N/A",
				UnloadTime:   "N/A",
			})
		}
		candidates = append(candidates, RouteCandidateDTO{Legs: legs})
	}

	return candidates
}

func (f *bookingServiceFacade) ListShippingLocations() []locationDTO {
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

func (f *bookingServiceFacade) ListAllCargos() []cargoDTO {
	cargos := f.cargoRepository.FindAll()
	dtos := make([]cargoDTO, len(cargos))

	for i, c := range cargos {
		dtos[i] = assemble(c)
	}

	return dtos
}

func NewBookingServiceFacade(cargoRepository cargo.CargoRepository, locationRepository location.LocationRepository, bookingService application.BookingService) BookingServiceFacade {
	return &bookingServiceFacade{cargoRepository, locationRepository, bookingService}
}

func assemble(c cargo.Cargo) cargoDTO {
	eta := time.Date(2009, time.March, 12, 12, 0, 0, 0, time.UTC)
	dto := cargoDTO{
		TrackingId:           string(c.TrackingId),
		StatusText:           fmt.Sprintf("%s %s", cargo.InPort, c.Origin.Name),
		Origin:               string(c.Origin.UNLocode),
		Destination:          string(c.RouteSpecification.Destination.UNLocode),
		ETA:                  eta.Format(time.RFC3339),
		NextExpectedActivity: "Next expected activity is to load cargo onto voyage 0200T in New York",
		Misrouted:            c.Delivery.RoutingStatus == cargo.Misrouted,
		Routed:               !c.Itinerary.IsEmpty(),
		ArrivalDeadline:      c.ArrivalDeadline.Format(time.RFC3339),
	}

	legs := make([]legDTO, 0)
	for _, l := range c.Itinerary.Legs {
		legs = append(legs, legDTO{
			VoyageNumber: string(l.VoyageNumber),
			From:         string(l.LoadLocation.UNLocode),
			To:           string(l.UnloadLocation.UNLocode),
		})
	}
	dto.Legs = legs

	// TODO: Query real handling history from handlingEventRepository
	dto.Events = make([]eventDTO, 3)
	dto.Events[0] = eventDTO{Description: "Received in Hongkong, at 3/1/09 12:00 AM.", Expected: true}
	dto.Events[1] = eventDTO{Description: "Loaded onto voyage 0100S in Hongkong, at 3/2/09 12:00 AM."}
	dto.Events[2] = eventDTO{Description: "Unloaded off voyage 0100S in New York, at 3/5/09 12:00 AM."}

	return dto
}
