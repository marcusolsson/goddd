package cargo

import (
	"container/list"
	"errors"
	"reflect"
	"time"

	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/shared"
)

// A HandlingEvent is used to register the event when, for instance, a
// cargo is unloaded from a carrier at a some location at a given
// time.
type HandlingEvent struct {
	TrackingId
	Type HandlingEventType
	location.Location
}

func (e HandlingEvent) SameValue(v shared.ValueObject) bool {
	return reflect.DeepEqual(e, v.(HandlingEvent))
}

// HandlingEventType either requires or prohibits a carrier movement
// association, it's never optional.
type HandlingEventType int

const (
	NotHandled HandlingEventType = iota
	Load
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

func (h HandlingHistory) SameValue(v shared.ValueObject) bool {
	return reflect.DeepEqual(h, v.(HandlingHistory))
}

// HandlingEventRepository
type HandlingEventRepository interface {
	Store(e HandlingEvent)
}

type HandlingEventRepositoryInMem struct {
}

func (r *HandlingEventRepositoryInMem) Store(e HandlingEvent) {
}

// HandlingEventFactory creates handling events.
type HandlingEventFactory struct {
	CargoRepository
}

var ErrCannotCreateHandlingEvent = errors.New("Cannot create handling event")

func (f *HandlingEventFactory) CreateHandlingEvent(registrationTime time.Time, completionTime time.Time, trackingId TrackingId,
	voyageNumber string, unLocode location.UNLocode, eventType HandlingEventType) (HandlingEvent, error) {
	_, err := f.CargoRepository.Find(trackingId)

	if err != nil {
		return HandlingEvent{}, ErrCannotCreateHandlingEvent
	}

	return HandlingEvent{}, nil
}
