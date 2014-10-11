package cargo

import (
	"errors"
	"reflect"
	"time"

	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/shared"
	"github.com/marcusolsson/goddd/domain/voyage"
)

type HandlingEvent struct {
	TrackingId
	Type     HandlingEventType
	Location location.UNLocode
	voyage.VoyageNumber
}

func (e HandlingEvent) SameValue(v shared.ValueObject) bool {
	return reflect.DeepEqual(e, v.(HandlingEvent))
}

type HandlingEventType int

const (
	NotHandled HandlingEventType = iota
	Load
	Unload
	Receive
	Claim
	Customs
)

type HandlingHistory struct {
	HandlingEvents []HandlingEvent
}

func (h HandlingHistory) MostRecentlyCompletedEvent() (HandlingEvent, error) {
	if len(h.HandlingEvents) == 0 {
		return HandlingEvent{}, errors.New("Delivery history is empty")
	}

	return h.HandlingEvents[len(h.HandlingEvents)-1], nil
}

func (h HandlingHistory) SameValue(v shared.ValueObject) bool {
	return reflect.DeepEqual(h, v.(HandlingHistory))
}

type HandlingEventRepository interface {
	Store(e HandlingEvent)
	QueryHandlingHistory(TrackingId) HandlingHistory
}

type HandlingEventFactory struct {
	CargoRepository
}

var ErrCannotCreateHandlingEvent = errors.New("Cannot create handling event")

func (f *HandlingEventFactory) CreateHandlingEvent(registrationTime time.Time, completionTime time.Time, trackingId TrackingId,
	voyageNumber voyage.VoyageNumber, unLocode location.UNLocode, eventType HandlingEventType) (HandlingEvent, error) {
	_, err := f.CargoRepository.Find(trackingId)

	if err != nil {
		return HandlingEvent{}, ErrCannotCreateHandlingEvent
	}

	return HandlingEvent{}, nil
}
