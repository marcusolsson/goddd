package tracking

import (
	"testing"

	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
)

type fakeCargos struct {
}

func (r *fakeCargos) Store(cargo *cargo.Cargo) error {
	return nil
}
func (r *fakeCargos) Find(trackingID cargo.TrackingID) (*cargo.Cargo, error) {
	return cargo.New("FTL456", cargo.RouteSpecification{
		Origin:      location.AUMEL,
		Destination: location.SESTO,
	}), nil
}
func (r *fakeCargos) FindAll() []*cargo.Cargo {
	return []*cargo.Cargo{}
}

type fakeHandlingEvents struct {
}

func (r *fakeHandlingEvents) Store(e cargo.HandlingEvent) {
}
func (r *fakeHandlingEvents) QueryHandlingHistory(cargo.TrackingID) cargo.HandlingHistory {
	return cargo.HandlingHistory{}
}

func TestTrack(t *testing.T) {
	cargos := &fakeCargos{}
	handlingEvents := &fakeHandlingEvents{}
	ts := NewService(cargos, handlingEvents)

	c, err := ts.Track("FTL456")
	if err != nil {
		t.Fatal(err)
	}

	if c.TrackingID != "FTL456" {
		t.Errorf("c.TrackingID = %v; want = %v", c.TrackingID, "ABC123")
	}
	if c.Origin != "AUMEL" {
		t.Errorf("c.Origin = %v; want = %v", c.Destination, "AUMEL")
	}
	if c.Destination != "SESTO" {
		t.Errorf("c.Destination = %v; want = %v", c.Destination, "SESTO")
	}
	if c.StatusText != "Not received" {
		t.Errorf("c.StatusText = %v; want = %v", c.StatusText, "Not received")
	}
}
