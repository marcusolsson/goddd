package handling

import (
	"testing"
	"time"

	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/mock"
	"github.com/marcusolsson/goddd/voyage"
)

type stubEventHandler struct {
	events []interface{}
}

func (h *stubEventHandler) CargoWasHandled(e cargo.HandlingEvent) {
	h.events = append(h.events, e)
}

func TestRegisterHandlingEvent(t *testing.T) {
	var cargos mock.CargoRepository
	cargos.StoreFn = func(c *cargo.Cargo) error {
		return nil
	}
	cargos.FindFn = func(id cargo.TrackingID) (*cargo.Cargo, error) {
		if id == "no_such_id" {
			return nil, cargo.ErrUnknown
		}
		return new(cargo.Cargo), nil
	}

	var voyages mock.VoyageRepository
	voyages.FindFn = func(n voyage.Number) (*voyage.Voyage, error) {
		return new(voyage.Voyage), nil
	}

	var locations mock.LocationRepository
	locations.FindFn = func(l location.UNLocode) (*location.Location, error) {
		return nil, nil
	}

	var events mock.HandlingEventRepository
	events.StoreFn = func(e cargo.HandlingEvent) {}

	eh := &stubEventHandler{events: make([]interface{}, 0)}
	ef := cargo.HandlingEventFactory{
		CargoRepository:    &cargos,
		VoyageRepository:   &voyages,
		LocationRepository: &locations,
	}

	s := NewService(&events, ef, eh)

	var (
		completed = time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC)
		id        = cargo.TrackingID("ABC123")
		voyage    = voyage.Number("V100")
	)

	var err error

	err = cargos.Store(cargo.New(id, cargo.RouteSpecification{}))
	if err != nil {
		t.Fatal(err)
	}

	err = s.RegisterHandlingEvent(completed, id, voyage, location.SESTO, cargo.Load)
	if err != nil {
		t.Fatal(err)
	}

	err = s.RegisterHandlingEvent(completed, "no_such_id", voyage, location.SESTO, cargo.Load)
	if err != cargo.ErrUnknown {
		t.Errorf("err = %s; want = %s", err, cargo.ErrUnknown)
	}

	if len(eh.events) != 1 {
		t.Errorf("len(eh.events) = %d; want = %d", len(eh.events), 1)
	}
}
