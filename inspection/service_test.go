package inspection

import (
	"testing"

	shipping "github.com/marcusolsson/goddd"
)

type stubEventHandler struct {
	events []interface{}
}

func (h *stubEventHandler) CargoWasMisdirected(c *shipping.Cargo) {
	h.events = append(h.events, c)
}

func (h *stubEventHandler) CargoHasArrived(c *shipping.Cargo) {
	h.events = append(h.events, c)
}

func TestInspectMisdirectedCargo(t *testing.T) {
	var cargos mockCargoRepository

	events := mockHandlingEventRepository{
		events: make(map[shipping.TrackingID][]shipping.HandlingEvent),
	}

	handler := stubEventHandler{make([]interface{}, 0)}

	s := NewService(&cargos, &events, &handler)

	id := shipping.TrackingID("ABC123")
	c := shipping.NewCargo(id, shipping.RouteSpecification{
		Origin:      shipping.SESTO,
		Destination: shipping.CNHKG,
	})

	voyage := shipping.VoyageNumber("001A")

	c.AssignToRoute(shipping.Itinerary{Legs: []shipping.Leg{
		{VoyageNumber: voyage, LoadLocation: shipping.SESTO, UnloadLocation: shipping.AUMEL},
		{VoyageNumber: voyage, LoadLocation: shipping.AUMEL, UnloadLocation: shipping.CNHKG},
	}})

	if err := cargos.Store(c); err != nil {
		t.Fatal(err)
	}

	storeEvent(&events, id, voyage, shipping.Receive, shipping.SESTO)
	storeEvent(&events, id, voyage, shipping.Load, shipping.SESTO)
	storeEvent(&events, id, voyage, shipping.Unload, shipping.USNYC)

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
		events: make(map[shipping.TrackingID][]shipping.HandlingEvent),
	}

	handler := stubEventHandler{make([]interface{}, 0)}

	s := &service{
		cargos:  &cargos,
		events:  &events,
		handler: &handler,
	}

	id := shipping.TrackingID("ABC123")
	unloadedCargo := shipping.NewCargo(id, shipping.RouteSpecification{
		Origin:      shipping.SESTO,
		Destination: shipping.CNHKG,
	})

	var voyage shipping.VoyageNumber = "001A"

	unloadedCargo.AssignToRoute(shipping.Itinerary{Legs: []shipping.Leg{
		{VoyageNumber: voyage, LoadLocation: shipping.SESTO, UnloadLocation: shipping.AUMEL},
		{VoyageNumber: voyage, LoadLocation: shipping.AUMEL, UnloadLocation: shipping.CNHKG},
	}})

	cargos.Store(unloadedCargo)

	storeEvent(&events, id, voyage, shipping.Receive, shipping.SESTO)
	storeEvent(&events, id, voyage, shipping.Load, shipping.SESTO)
	storeEvent(&events, id, voyage, shipping.Unload, shipping.AUMEL)
	storeEvent(&events, id, voyage, shipping.Load, shipping.AUMEL)
	storeEvent(&events, id, voyage, shipping.Unload, shipping.CNHKG)

	if len(handler.events) != 0 {
		t.Errorf("len(handler.events) = %d; want = %d", len(handler.events), 0)
	}

	s.InspectCargo(id)

	if len(handler.events) != 1 {
		t.Errorf("len(handler.events) = %d; want = %d", len(handler.events), 1)
	}
}

func storeEvent(r shipping.HandlingEventRepository, id shipping.TrackingID, voyageNumber shipping.VoyageNumber, typ shipping.HandlingEventType, loc shipping.UNLocode) {
	e := shipping.HandlingEvent{
		TrackingID: id,
		Activity: shipping.HandlingActivity{
			VoyageNumber: voyageNumber,
			Type:         typ,
			Location:     loc,
		},
	}

	r.Store(e)
}

type mockCargoRepository struct {
	cargo *shipping.Cargo
}

func (r *mockCargoRepository) Store(c *shipping.Cargo) error {
	r.cargo = c
	return nil
}

func (r *mockCargoRepository) Find(id shipping.TrackingID) (*shipping.Cargo, error) {
	if r.cargo != nil {
		return r.cargo, nil
	}
	return nil, shipping.ErrUnknownCargo
}

func (r *mockCargoRepository) FindAll() []*shipping.Cargo {
	return []*shipping.Cargo{r.cargo}
}

type mockHandlingEventRepository struct {
	events map[shipping.TrackingID][]shipping.HandlingEvent
}

func (r *mockHandlingEventRepository) Store(e shipping.HandlingEvent) {
	if _, ok := r.events[e.TrackingID]; !ok {
		r.events[e.TrackingID] = make([]shipping.HandlingEvent, 0)
	}
	r.events[e.TrackingID] = append(r.events[e.TrackingID], e)
}

func (r *mockHandlingEventRepository) QueryHandlingHistory(id shipping.TrackingID) shipping.HandlingHistory {
	return shipping.HandlingHistory{HandlingEvents: r.events[id]}
}
