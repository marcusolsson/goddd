package application

import (
	"time"

	"github.com/marcusolsson/goddd/domain/cargo"
	"github.com/marcusolsson/goddd/domain/location"
)

type HandlingEventService interface {
	RegisterHandlingEvent(completionTime time.Time, trackingId cargo.TrackingId, voyageNumber string,
		unLocode location.UNLocode, eventType cargo.HandlingEventType) error
}

type handlingEventService struct {
	handlingEventRepository cargo.HandlingEventRepository
	handlingEventFactory    cargo.HandlingEventFactory
}

func (s *handlingEventService) RegisterHandlingEvent(completionTime time.Time, trackingId cargo.TrackingId, voyageNumber string,
	unLocode location.UNLocode, eventType cargo.HandlingEventType) error {

	registrationTime := time.Now()
	event, _ := s.handlingEventFactory.CreateHandlingEvent(registrationTime, completionTime, trackingId, voyageNumber, unLocode, eventType)

	s.handlingEventRepository.Store(event)

	return nil
}
