package inspection

import (
	"testing"

	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/voyage"
)

type stubEventHandler struct {
	events []interface{}
}

func (h *stubEventHandler) CargoWasMisdirected(c *cargo.Cargo) {
	h.events = append(h.events, c)
}

func (h *stubEventHandler) CargoHasArrived(c *cargo.Cargo) {
	h.events = append(h.events, c)
}

func TestInspectMisdirectedCargo(t *testing.T) {
	var cargos mockCargoRepository

	events := mockHandlingEventRepository{
		events: make(map[cargo.TrackingID][]cargo.HandlingEvent),
	}

	handler := stubEventHandler{make([]interface{}, 0)}

	s := NewService(&cargos, &events, &handler)

	id := cargo.TrackingID("ABC123")
	c := cargo.New(id, cargo.RouteSpecification{
		Origin:      location.SESTO,
		Destination: location.CNHKG,
	})

	voyage := voyage.Number("001A")

	c.AssignToRoute(cargo.Itinerary{Legs: []cargo.Leg{
		{VoyageNumber: voyage, LoadLocation: location.SESTO, UnloadLocation: location.AUMEL},
		{VoyageNumber: voyage, LoadLocation: location.AUMEL, UnloadLocation: location.CNHKG},
	}})

	if err := cargos.Store(c); err != nil {
		t.Fatal(err)
	}

	storeEvent(&events, id, voyage, cargo.Receive, location.SESTO)
	storeEvent(&events, id, voyage, cargo.Load, location.SESTO)
	storeEvent(&events, id, voyage, cargo.Unload, location.USNYC)

	if len(handler.events) != 0 {
		t.Errorf("no events should be handled")
	}

	s.InspectCargo(id)

	if len(handler.events) != 1 {
		t.Errorf("1 event should be handled")
	}

	s.InspectCargo("no_such_id")

	// no events was published
	if len(handler.events) != 1 {
		t.Errorf("len(handler.events) = %d; want = %d", len(handler.events), 1)
	}
}

func TestInspectUnloadedCargo(t *testing.T) {
	var cargos mockCargoRepository

	events := mockHandlingEventRepository{
		events: make(map[cargo.TrackingID][]cargo.HandlingEvent),
	}

	handler := stubEventHandler{make([]interface{}, 0)}

	s := &service{
		cargos:  &cargos,
		events:  &events,
		handler: &handler,
	}

	id := cargo.TrackingID("ABC123")
	unloadedCargo := cargo.New(id, cargo.RouteSpecification{
		Origin:      location.SESTO,
		Destination: location.CNHKG,
	})

	var voyage voyage.Number = "001A"

	unloadedCargo.AssignToRoute(cargo.Itinerary{Legs: []cargo.Leg{
		{VoyageNumber: voyage, LoadLocation: location.SESTO, UnloadLocation: location.AUMEL},
		{VoyageNumber: voyage, LoadLocation: location.AUMEL, UnloadLocation: location.CNHKG},
	}})

	cargos.Store(unloadedCargo)

	storeEvent(&events, id, voyage, cargo.Receive, location.SESTO)
	storeEvent(&events, id, voyage, cargo.Load, location.SESTO)
	storeEvent(&events, id, voyage, cargo.Unload, location.AUMEL)
	storeEvent(&events, id, voyage, cargo.Load, location.AUMEL)
	storeEvent(&events, id, voyage, cargo.Unload, location.CNHKG)

	if len(handler.events) != 0 {
		t.Errorf("len(handler.events) = %d; want = %d", len(handler.events), 0)
	}

	s.InspectCargo(id)

	if len(handler.events) != 1 {
		t.Errorf("len(handler.events) = %d; want = %d", len(handler.events), 1)
	}
}

func storeEvent(r cargo.HandlingEventRepository, id cargo.TrackingID, voyageNumber voyage.Number, typ cargo.HandlingEventType, loc location.UNLocode) {
	e := cargo.HandlingEvent{
		TrackingID: id,
		Activity: cargo.HandlingActivity{
			VoyageNumber: voyageNumber,
			Type:         typ,
			Location:     loc,
		},
	}

	r.Store(e)
}

type mockCargoRepository struct {
	cargo *cargo.Cargo
}

func (r *mockCargoRepository) Store(c *cargo.Cargo) error {
	r.cargo = c
	return nil
}

func (r *mockCargoRepository) Find(id cargo.TrackingID) (*cargo.Cargo, error) {
	if r.cargo != nil {
		return r.cargo, nil
	}
	return nil, cargo.ErrUnknown
}

func (r *mockCargoRepository) FindAll() []*cargo.Cargo {
	return []*cargo.Cargo{r.cargo}
}

type mockHandlingEventRepository struct {
	events map[cargo.TrackingID][]cargo.HandlingEvent
}

func (r *mockHandlingEventRepository) Store(e cargo.HandlingEvent) {
	if _, ok := r.events[e.TrackingID]; !ok {
		r.events[e.TrackingID] = make([]cargo.HandlingEvent, 0)
	}
	r.events[e.TrackingID] = append(r.events[e.TrackingID], e)
}

func (r *mockHandlingEventRepository) QueryHandlingHistory(id cargo.TrackingID) cargo.HandlingHistory {
	return cargo.HandlingHistory{HandlingEvents: r.events[id]}
}
