// Package handling provides the use-case for registering incidents. Used by
// views facing the people handling the cargo along its route.
package handling

import (
	"errors"
	"time"

	shipping "github.com/marcusolsson/goddd"
	"github.com/marcusolsson/goddd/inspection"
)

// ErrInvalidArgument is returned when one or more arguments are invalid.
var ErrInvalidArgument = errors.New("invalid argument")

// EventHandler provides a means of subscribing to registered handling events.
type EventHandler interface {
	CargoWasHandled(shipping.HandlingEvent)
}

// Service provides handling operations.
type Service interface {
	// RegisterHandlingEvent registers a handling event in the system, and
	// notifies interested parties that a cargo has been handled.
	RegisterHandlingEvent(completed time.Time, id shipping.TrackingID, voyageNumber shipping.VoyageNumber,
		unLocode shipping.UNLocode, eventType shipping.HandlingEventType) error
}

type service struct {
	handlingEventRepository shipping.HandlingEventRepository
	handlingEventFactory    shipping.HandlingEventFactory
	handlingEventHandler    EventHandler
}

// NewService creates a handling event service with necessary dependencies.
func NewService(r shipping.HandlingEventRepository, f shipping.HandlingEventFactory, h EventHandler) Service {
	return &service{
		handlingEventRepository: r,
		handlingEventFactory:    f,
		handlingEventHandler:    h,
	}
}

func (s *service) RegisterHandlingEvent(completed time.Time, id shipping.TrackingID, voyageNumber shipping.VoyageNumber,
	loc shipping.UNLocode, eventType shipping.HandlingEventType) error {
	if completed.IsZero() || id == "" || loc == "" || eventType == shipping.NotHandled {
		return ErrInvalidArgument
	}

	e, err := s.handlingEventFactory.MakeHandlingEvent(time.Now(), completed, id, voyageNumber, loc, eventType)
	if err != nil {
		return err
	}

	s.handlingEventRepository.Store(e)
	s.handlingEventHandler.CargoWasHandled(e)

	return nil
}

type handlingEventHandler struct {
	InspectionService inspection.Service
}

func (h *handlingEventHandler) CargoWasHandled(event shipping.HandlingEvent) {
	h.InspectionService.InspectCargo(event.TrackingID)
}

// NewEventHandler returns a new instance of a EventHandler.
func NewEventHandler(s inspection.Service) EventHandler {
	return &handlingEventHandler{
		InspectionService: s,
	}
}
