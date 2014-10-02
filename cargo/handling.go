package cargo

import (
	"container/list"
	"errors"
	"time"

	"bitbucket.org/marcus_olsson/goddd/location"
)

// A HandlingEvent is used to register the event when, for instance, a
// cargo is unloaded from a carrier at a some location at a given
// time.
type HandlingEvent struct {
	TrackingId
	Type HandlingEventType
	location.Location
}

// HandlingEventType either requires or prohibits a carrier movement
// association, it's never optional.
type HandlingEventType int

const (
	Load HandlingEventType = iota
	Unload
	Receive
	Claim
	Customs
)

// HandlingHistory is the handling history of a cargo.
type HandlingHistory struct {
	HandlingEvents *list.List
}

func (h HandlingHistory) MostRecentlyCompletedEvent() (HandlingEvent, error) {
	if h.HandlingEvents.Len() == 0 {
		return HandlingEvent{}, errors.New("Delivery history is empty")
	}

	return h.HandlingEvents.Back().Value.(HandlingEvent), nil
}

type HandlingEventRepository interface {
	// Stores a (new) handling event.
	Store(e HandlingEvent)
}

type handlingEventRepository struct {
	events *list.List
}

func (r *handlingEventRepository) Store(e HandlingEvent) {

}

// HandlingEventFactory creates handling events.
type HandlingEventFactory interface {
	CreateHandlingEvent(trackingId TrackingId) (HandlingEvent, error)
}

var (
	ErrCannotCreateHandlingEvent = errors.New("Cannot create handling event")
)

type handlingEventFactory struct {
	cargoRepository CargoRepository
}

func (f *handlingEventFactory) CreateHandlingEvent(trackingId TrackingId) (HandlingEvent, error) {
	_, err := f.cargoRepository.Find(trackingId)

	if err != nil {
		return HandlingEvent{}, ErrCannotCreateHandlingEvent
	}

	return HandlingEvent{}, nil
}

type HandlingEventService interface {
	// Registers a handling event in the system, and notifies
	// interested parties that a cargo has been handled.
	RegisterHandlingEvent(completionTime time.Time, trackingId TrackingId, unLocode location.UNLocode, eventType HandlingEventType) error
}

type handlingEventService struct {
	repository HandlingEventRepository
	factory    HandlingEventFactory
}

func (s *handlingEventService) RegisterHandlingEvent(completionTime time.Time, trackingId TrackingId, unLocode location.UNLocode, eventType HandlingEventType) error {
	e, _ := s.factory.CreateHandlingEvent(trackingId)

	s.repository.Store(e)

	return nil
}

func NewHandlingEventService(r HandlingEventRepository, f HandlingEventFactory) HandlingEventService {
	return &handlingEventService{
		repository: r,
		factory:    f,
	}
}
