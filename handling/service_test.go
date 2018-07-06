package handling

import (
	"testing"
	"time"

	shipping "github.com/marcusolsson/goddd"
	"github.com/marcusolsson/goddd/mock"
)

type stubEventHandler struct {
	events []interface{}
}

func (h *stubEventHandler) CargoWasHandled(e shipping.HandlingEvent) {
	h.events = append(h.events, e)
}

func TestRegisterHandlingEvent(t *testing.T) {
	var cargos mock.CargoRepository
	cargos.StoreFn = func(c *shipping.Cargo) error {
		return nil
	}
	cargos.FindFn = func(id shipping.TrackingID) (*shipping.Cargo, error) {
		if id == "no_such_id" {
			return nil, shipping.ErrUnknownCargo
		}
		return new(shipping.Cargo), nil
	}

	var voyages mock.VoyageRepository
	voyages.FindFn = func(n shipping.VoyageNumber) (*shipping.Voyage, error) {
		return new(shipping.Voyage), nil
	}

	var locations mock.LocationRepository
	locations.FindFn = func(l shipping.UNLocode) (*shipping.Location, error) {
		return nil, nil
	}

	var events mock.HandlingEventRepository
	events.StoreFn = func(e shipping.HandlingEvent) {}

	eh := &stubEventHandler{events: make([]interface{}, 0)}
	ef := shipping.HandlingEventFactory{
		CargoRepository:    &cargos,
		VoyageRepository:   &voyages,
		LocationRepository: &locations,
	}

	s := NewService(&events, ef, eh)

	var (
		completed = time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC)
		id        = shipping.TrackingID("ABC123")
		voyage    = shipping.VoyageNumber("V100")
	)

	var err error

	err = cargos.Store(shipping.NewCargo(id, shipping.RouteSpecification{}))
	if err != nil {
		t.Fatal(err)
	}

	err = s.RegisterHandlingEvent(completed, id, voyage, shipping.SESTO, shipping.Load)
	if err != nil {
		t.Fatal(err)
	}

	err = s.RegisterHandlingEvent(completed, "no_such_id", voyage, shipping.SESTO, shipping.Load)
	if err != shipping.ErrUnknownCargo {
		t.Errorf("err = %s; want = %s", err, shipping.ErrUnknownCargo)
	}

	if len(eh.events) != 1 {
		t.Errorf("len(eh.events) = %d; want = %d", len(eh.events), 1)
	}
}
