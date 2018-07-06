package booking

import (
	"testing"
	"time"

	shipping "github.com/marcusolsson/goddd"
	"github.com/marcusolsson/goddd/mock"
)

func TestBookNewCargo(t *testing.T) {
	var (
		origin      = shipping.SESTO
		destination = shipping.AUMEL
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

func (s *stubRoutingService) FetchRoutesForSpecification(rs shipping.RouteSpecification) []shipping.Itinerary {
	legs := []shipping.Leg{
		{LoadLocation: rs.Origin, UnloadLocation: rs.Destination},
	}

	return []shipping.Itinerary{
		{Legs: legs},
	}
}

func TestRequestPossibleRoutesForCargo(t *testing.T) {
	var (
		origin      = shipping.SESTO
		destination = shipping.AUMEL
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
		origin      = shipping.SESTO
		destination = shipping.AUMEL
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

	if err := s.AssignCargoToRoute("no_such_id", shipping.Itinerary{}); err != ErrInvalidArgument {
		t.Errorf("err = %s; want = %s", err, ErrInvalidArgument)
	}
}

func TestChangeCargoDestination(t *testing.T) {
	var cargos mockCargoRepository
	var locations mock.LocationRepository

	locations.FindFn = func(loc shipping.UNLocode) (*shipping.Location, error) {
		if loc != shipping.AUMEL {
			return nil, shipping.ErrUnknownLocation
		}
		return shipping.Melbourne, nil
	}

	var rs stubRoutingService

	s := NewService(&cargos, &locations, nil, &rs)

	c := shipping.NewCargo("ABC", shipping.RouteSpecification{
		Origin:          shipping.SESTO,
		Destination:     shipping.CNHKG,
		ArrivalDeadline: time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC),
	})

	if err := s.ChangeDestination("no_such_id", shipping.SESTO); err != shipping.ErrUnknownCargo {
		t.Errorf("err = %s; want = %s", err, shipping.ErrUnknownCargo)
	}

	if err := cargos.Store(c); err != nil {
		t.Fatal(err)
	}

	if err := s.ChangeDestination(c.TrackingID, "no_such_unlocode"); err != shipping.ErrUnknownLocation {
		t.Errorf("err = %s; want = %s", err, shipping.ErrUnknownLocation)
	}

	if c.RouteSpecification.Destination != shipping.CNHKG {
		t.Errorf("c.RouteSpecification.Destination = %s; want = %s",
			c.RouteSpecification.Destination, shipping.CNHKG)
	}

	if err := s.ChangeDestination(c.TrackingID, shipping.AUMEL); err != nil {
		t.Fatal(err)
	}

	uc, err := cargos.Find(c.TrackingID)
	if err != nil {
		t.Fatal(err)
	}

	if uc.RouteSpecification.Destination != shipping.AUMEL {
		t.Errorf("uc.RouteSpecification.Destination = %s; want = %s",
			uc.RouteSpecification.Destination, shipping.AUMEL)
	}
}

func TestLoadCargo(t *testing.T) {
	deadline := time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC)

	var cargos mock.CargoRepository
	cargos.FindFn = func(id shipping.TrackingID) (*shipping.Cargo, error) {
		return &shipping.Cargo{
			TrackingID: "test_id",
			Origin:     shipping.SESTO,
			RouteSpecification: shipping.RouteSpecification{
				Origin:          shipping.SESTO,
				Destination:     shipping.AUMEL,
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
