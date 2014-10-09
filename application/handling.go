package application

import (
	"time"

	"github.com/marcusolsson/goddd/domain/cargo"
	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/voyage"
)

type HandlingEventService interface {
	RegisterHandlingEvent(completionTime time.Time, trackingId cargo.TrackingId, voyageNumber voyage.VoyageNumber,
		unLocode location.UNLocode, eventType cargo.HandlingEventType) error
}

type handlingEventService struct {
	handlingEventRepository cargo.HandlingEventRepository
	handlingEventFactory    cargo.HandlingEventFactory
	eventHandler            EventHandler
}

func (s *handlingEventService) RegisterHandlingEvent(completionTime time.Time, trackingId cargo.TrackingId, voyageNumber voyage.VoyageNumber,
	unLocode location.UNLocode, eventType cargo.HandlingEventType) error {

	registrationTime := time.Now()
	event, err := s.handlingEventFactory.CreateHandlingEvent(registrationTime, completionTime, trackingId, voyageNumber, unLocode, eventType)

	if err != nil {
		return err
	}

	s.handlingEventRepository.Store(event)

	s.eventHandler.CargoWasHandled(event)

	return nil
}
