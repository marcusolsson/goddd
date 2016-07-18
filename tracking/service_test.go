package tracking

import (
	"testing"

	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/mock"
)

func TestTrack(t *testing.T) {
	cargos := &mock.CargoRepository{
		FindFn: func(id cargo.TrackingID) (*cargo.Cargo, error) {
			return cargo.New("FTL456", cargo.RouteSpecification{
				Origin:      location.AUMEL,
				Destination: location.SESTO,
			}), nil
		},
	}

	handlingEvents := &mock.HandlingEventRepository{
		QueryHandlingHistoryFn: func(id cargo.TrackingID) cargo.HandlingHistory {
			return cargo.HandlingHistory{}
		},
	}

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
