package application

import (
	"time"

	"github.com/marcusolsson/goddd/domain/cargo"
	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/voyage"
)

type HandlingEventService interface {
	// RegisterHandlingEvent registers a handling event in the system, and
	// notifies interested parties that a cargo has been handled.
	RegisterHandlingEvent(completionTime time.Time, trackingID cargo.TrackingID, voyageNumber voyage.Number,
		unLocode location.UNLocode, eventType cargo.HandlingEventType) error
}

type handlingEventService struct {
	handlingEventRepository cargo.HandlingEventRepository
	handlingEventFactory    cargo.HandlingEventFactory
	handlingEventHandler    HandlingEventHandler
}

func (s *handlingEventService) RegisterHandlingEvent(completionTime time.Time, trackingID cargo.TrackingID, voyageNumber voyage.Number,
	unLocode location.UNLocode, eventType cargo.HandlingEventType) error {

	registrationTime := time.Now()

	event, err := s.handlingEventFactory.CreateHandlingEvent(registrationTime, completionTime, trackingID, voyageNumber, unLocode, eventType)

	if err != nil {
		return err
	}

	s.handlingEventRepository.Store(event)

	s.handlingEventHandler.CargoWasHandled(event)

	return nil
}

// NewHandlingEventService creates a handling event service with necessary
// dependencies.
func NewHandlingEventService(r cargo.HandlingEventRepository, f cargo.HandlingEventFactory, h HandlingEventHandler) HandlingEventService {
	return &handlingEventService{
		handlingEventRepository: r,
		handlingEventFactory:    f,
		handlingEventHandler:    h,
	}
}
