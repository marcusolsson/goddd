package booking

import (
	"testing"
	"time"

	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/mock"
)

func TestBookNewCargo(t *testing.T) {
	var (
		origin      = location.SESTO
		destination = location.AUMEL
		deadline    = time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC)
	)

	var cargos mockCargoRepository

	s := NewService(&cargos, nil, nil, nil)

	id, err := s.BookNewCargo(origin, destination, deadline)
	if err != nil {
		t.Fatal(err)
	}

	c, err := cargos.Find(id)
	if err != nil {
		t.Fatal(err)
	}

	if c.TrackingID != id {
		t.Errorf("c.TrackingID = %s; want = %s", c.TrackingID, id)
	}
	if c.Origin != origin {
		t.Errorf("c.Origin = %s; want = %s", c.Origin, origin)
	}
	if c.RouteSpecification.Destination != destination {
		t.Errorf("c.RouteSpecification.Destination = %s; want = %s",
			c.RouteSpecification.Destination, destination)
	}
	if c.RouteSpecification.ArrivalDeadline != deadline {
		t.Errorf("c.RouteSpecification.ArrivalDeadline = %s; want = %s",
			c.RouteSpecification.ArrivalDeadline, deadline)
	}
}

type stubRoutingService struct{}

func (s *stubRoutingService) FetchRoutesForSpecification(rs cargo.RouteSpecification) []cargo.Itinerary {
	legs := []cargo.Leg{
		{LoadLocation: rs.Origin, UnloadLocation: rs.Destination},
	}

	return []cargo.Itinerary{
		{Legs: legs},
	}
}

func TestRequestPossibleRoutesForCargo(t *testing.T) {
	var (
		origin      = location.SESTO
		destination = location.AUMEL
		deadline    = time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC)
	)

	var cargos mockCargoRepository

	var rs stubRoutingService

	s := NewService(&cargos, nil, nil, &rs)

	r := s.RequestPossibleRoutesForCargo("no_such_id")

	if len(r) != 0 {
		t.Errorf("len(r) = %d; want = %d", len(r), 0)
	}

	id, err := s.BookNewCargo(origin, destination, deadline)
	if err != nil {
		t.Fatal(err)
	}

	i := s.RequestPossibleRoutesForCargo(id)

	if len(i) != 1 {
		t.Errorf("len(i) = %d; want = %d", len(i), 1)
	}
}

func TestAssignCargoToRoute(t *testing.T) {
	var cargos mockCargoRepository

	var rs stubRoutingService

	s := NewService(&cargos, nil, nil, &rs)

	var (
		origin      = location.SESTO
		destination = location.AUMEL
		deadline    = time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC)
	)

	id, err := s.BookNewCargo(origin, destination, deadline)
	if err != nil {
		t.Fatal(err)
	}

	i := s.RequestPossibleRoutesForCargo(id)

	if len(i) != 1 {
		t.Errorf("len(i) = %d; want = %d", len(i), 1)
	}

	if err := s.AssignCargoToRoute(id, i[0]); err != nil {
		t.Fatal(err)
	}

	if err := s.AssignCargoToRoute("no_such_id", cargo.Itinerary{}); err != ErrInvalidArgument {
		t.Errorf("err = %s; want = %s", err, ErrInvalidArgument)
	}
}

func TestChangeCargoDestination(t *testing.T) {
	var cargos mockCargoRepository
	var locations mock.LocationRepository

	locations.FindFn = func(loc location.UNLocode) (*location.Location, error) {
		if loc != location.AUMEL {
			return nil, location.ErrUnknown
		}
		return location.Melbourne, nil
	}

	var rs stubRoutingService

	s := NewService(&cargos, &locations, nil, &rs)

	c := cargo.New("ABC", cargo.RouteSpecification{
		Origin:          location.SESTO,
		Destination:     location.CNHKG,
		ArrivalDeadline: time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC),
	})

	if err := s.ChangeDestination("no_such_id", location.SESTO); err != cargo.ErrUnknown {
		t.Errorf("err = %s; want = %s", err, cargo.ErrUnknown)
	}

	if err := cargos.Store(c); err != nil {
		t.Fatal(err)
	}

	if err := s.ChangeDestination(c.TrackingID, "no_such_unlocode"); err != location.ErrUnknown {
		t.Errorf("err = %s; want = %s", err, location.ErrUnknown)
	}

	if c.RouteSpecification.Destination != location.CNHKG {
		t.Errorf("c.RouteSpecification.Destination = %s; want = %s",
			c.RouteSpecification.Destination, location.CNHKG)
	}

	if err := s.ChangeDestination(c.TrackingID, location.AUMEL); err != nil {
		t.Fatal(err)
	}

	uc, err := cargos.Find(c.TrackingID)
	if err != nil {
		t.Fatal(err)
	}

	if uc.RouteSpecification.Destination != location.AUMEL {
		t.Errorf("uc.RouteSpecification.Destination = %s; want = %s",
			uc.RouteSpecification.Destination, location.AUMEL)
	}
}

func TestLoadCargo(t *testing.T) {
	deadline := time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC)

	var cargos mock.CargoRepository
	cargos.FindFn = func(id cargo.TrackingID) (*cargo.Cargo, error) {
		return &cargo.Cargo{
			TrackingID: "test_id",
			Origin:     location.SESTO,
			RouteSpecification: cargo.RouteSpecification{
				Origin:          location.SESTO,
				Destination:     location.AUMEL,
				ArrivalDeadline: deadline,
			},
		}, nil
	}

	s := NewService(&cargos, nil, nil, nil)

	c, err := s.LoadCargo("test_id")
	if err != nil {
		t.Fatal(err)
	}

	if c.TrackingID != "test_id" {
		t.Errorf("c.TrackingID = %s; want = %s", c.TrackingID, "test_id")
	}
	if c.Origin != "SESTO" {
		t.Errorf("c.Origin = %s; want = %s", c.Origin, "SESTO")
	}
	if c.Destination != "AUMEL" {
		t.Errorf("c.Destination = %s; want = %s", c.Origin, "AUMEL")
	}
	if c.ArrivalDeadline != deadline {
		t.Errorf("c.ArrivalDeadline = %s; want = %s", c.ArrivalDeadline, deadline)
	}
	if c.Misrouted {
		t.Errorf("cargo should not be misrouted")
	}
	if c.Routed {
		t.Errorf("cargo should not have been routed")
	}
	if len(c.Legs) != 0 {
		t.Errorf("len(c.Legs) = %d; want = %d", len(c.Legs), 0)
	}
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
