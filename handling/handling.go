package handling

import (
	"time"

	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/inspection"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/voyage"
)

type EventHandler interface {
	CargoWasHandled(cargo.HandlingEvent)
}

type Service interface {
	// RegisterHandlingEvent registers a handling event in the system, and
	// notifies interested parties that a cargo has been handled.
	RegisterHandlingEvent(completionTime time.Time, trackingID cargo.TrackingID, voyageNumber voyage.Number,
		unLocode location.UNLocode, eventType cargo.HandlingEventType) error
}

type service struct {
	handlingEventRepository cargo.HandlingEventRepository
	handlingEventFactory    cargo.HandlingEventFactory
	handlingEventHandler    EventHandler
}

func (s *service) RegisterHandlingEvent(completionTime time.Time, trackingID cargo.TrackingID, voyageNumber voyage.Number,
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

// NewService creates a handling event service with necessary dependencies.
func NewService(r cargo.HandlingEventRepository, f cargo.HandlingEventFactory, h EventHandler) Service {
	return &service{
		handlingEventRepository: r,
		handlingEventFactory:    f,
		handlingEventHandler:    h,
	}
}

type handlingEventHandler struct {
	InspectionService inspection.Service
}

func (h *handlingEventHandler) CargoWasHandled(event cargo.HandlingEvent) {
	h.InspectionService.InspectCargo(event.TrackingID)
}

func NewEventHandler(s inspection.Service) *handlingEventHandler {
	return &handlingEventHandler{
		InspectionService: s,
	}
}
