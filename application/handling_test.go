package application

import (
	"time"

	"github.com/marcusolsson/goddd/domain/cargo"
	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/infrastructure"
	. "gopkg.in/check.v1"
)

func (s *S) TestRegisterHandlingEvent(c *C) {

	var (
		cargoRepository         = infrastructure.NewInMemCargoRepository()
		handlingEventRepository = &cargo.HandlingEventRepositoryInMem{}
		handlingEventFactory    = cargo.HandlingEventFactory{
			CargoRepository: cargoRepository,
		}
	)

	var service HandlingEventService = &handlingEventService{
		handlingEventRepository: handlingEventRepository,
		handlingEventFactory:    handlingEventFactory,
	}

	var (
		completionTime = time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC)
		trackingId     = cargo.TrackingId("ABC123")
		voyageNumber   = "CM001"
		unLocode       = location.Stockholm.UNLocode
		eventType      = cargo.Load
	)

	service.RegisterHandlingEvent(completionTime, trackingId, voyageNumber, unLocode, eventType)
}
