package handling

import (
	"testing"
	"time"

	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/repository"
	"github.com/marcusolsson/goddd/voyage"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

type stubEventHandler struct {
	handledEvents []interface{}
}

func (h *stubEventHandler) CargoWasHandled(event cargo.HandlingEvent) {
	h.handledEvents = append(h.handledEvents, event)
}

func (s *S) TestRegisterHandlingEvent(c *C) {
	var (
		cargoRepository         = repository.NewCargo()
		voyageRepository        = repository.NewVoyage()
		locationRepository      = repository.NewLocation()
		handlingEventRepository = repository.NewHandlingEvent()
	)

	var (
		handlingEventHandler = &stubEventHandler{make([]interface{}, 0)}
		handlingEventFactory = cargo.HandlingEventFactory{
			CargoRepository:    cargoRepository,
			VoyageRepository:   voyageRepository,
			LocationRepository: locationRepository,
		}
	)

	handlingEventService := NewService(handlingEventRepository, handlingEventFactory, handlingEventHandler)

	var (
		completionTime = time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC)
		trackingID     = cargo.TrackingID("ABC123")
		voyageNumber   = voyage.Number("V100")
		unLocode       = location.Stockholm.UNLocode
		eventType      = cargo.Load
	)

	cargoRepository.Store(cargo.New(trackingID, cargo.RouteSpecification{}))

	var err error

	err = handlingEventService.RegisterHandlingEvent(completionTime, trackingID, voyageNumber, unLocode, eventType)
	c.Check(err, IsNil)

	err = handlingEventService.RegisterHandlingEvent(completionTime, "no_such_id", voyageNumber, unLocode, eventType)
	c.Check(err, Not(IsNil))

	c.Check(len(handlingEventHandler.handledEvents), Equals, 1)
}
