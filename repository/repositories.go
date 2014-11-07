package repository

import (
	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/voyage"
)

type cargoRepositoryInMem struct {
	cargos map[cargo.TrackingID]cargo.Cargo
}

func (r *cargoRepositoryInMem) Store(c cargo.Cargo) error {
	r.cargos[c.TrackingID] = c

	return nil
}

func (r *cargoRepositoryInMem) Find(trackingID cargo.TrackingID) (cargo.Cargo, error) {

	if val, ok := r.cargos[trackingID]; ok {
		return val, nil
	}

	return cargo.Cargo{}, cargo.ErrUnknownCargo
}

func (r *cargoRepositoryInMem) FindAll() []cargo.Cargo {
	c := make([]cargo.Cargo, 0, len(r.cargos))
	for _, val := range r.cargos {
		c = append(c, val)
	}
	return c
}

func NewInMemCargoRepository() cargo.Repository {
	return &cargoRepositoryInMem{
		cargos: make(map[cargo.TrackingID]cargo.Cargo),
	}
}

type locationRepositoryInMem struct {
	locations map[location.UNLocode]location.Location
}

func (r *locationRepositoryInMem) Find(locode location.UNLocode) (location.Location, error) {
	if l, ok := r.locations[locode]; ok {
		return l, nil
	}

	return location.Location{}, location.ErrUnknownLocation
}

func (r *locationRepositoryInMem) FindAll() []location.Location {
	l := make([]location.Location, 0, len(r.locations))
	for _, val := range r.locations {
		l = append(l, val)
	}
	return l
}

func NewInMemLocationRepository() location.Repository {
	r := &locationRepositoryInMem{
		locations: make(map[location.UNLocode]location.Location),
	}

	r.locations[location.SESTO] = location.Stockholm
	r.locations[location.AUMEL] = location.Melbourne
	r.locations[location.CNHKG] = location.Hongkong
	r.locations[location.JNTKO] = location.Tokyo
	r.locations[location.NLRTM] = location.Rotterdam
	r.locations[location.DEHAM] = location.Hamburg

	return r
}

type voyageRepositoryInMem struct {
	voyages map[voyage.Number]voyage.Voyage
}

func (r *voyageRepositoryInMem) Find(voyageNumber voyage.Number) (voyage.Voyage, error) {
	if v, ok := r.voyages[voyageNumber]; ok {
		return v, nil
	}

	return voyage.Voyage{}, voyage.ErrUnknownVoyage
}

func NewInMemVoyageRepository() voyage.Repository {
	r := &voyageRepositoryInMem{
		voyages: make(map[voyage.Number]voyage.Voyage),
	}

	r.voyages[voyage.V100.Number] = *voyage.V100
	r.voyages[voyage.V300.Number] = *voyage.V300
	r.voyages[voyage.V400.Number] = *voyage.V400

	r.voyages[voyage.V0100S.Number] = *voyage.V0100S
	r.voyages[voyage.V0200T.Number] = *voyage.V0200T
	r.voyages[voyage.V0300A.Number] = *voyage.V0300A
	r.voyages[voyage.V0301S.Number] = *voyage.V0301S
	r.voyages[voyage.V0400S.Number] = *voyage.V0400S

	return r
}

type handlingEventRepositoryInMem struct {
	events map[cargo.TrackingID][]cargo.HandlingEvent
}

func (r *handlingEventRepositoryInMem) Store(e cargo.HandlingEvent) {
	// Make array if it's the first event with this tracking ID.
	if _, ok := r.events[e.TrackingID]; !ok {
		r.events[e.TrackingID] = make([]cargo.HandlingEvent, 0)
	}
	r.events[e.TrackingID] = append(r.events[e.TrackingID], e)
}

func (r *handlingEventRepositoryInMem) QueryHandlingHistory(trackingID cargo.TrackingID) cargo.HandlingHistory {
	return cargo.HandlingHistory{r.events[trackingID]}
}

func NewInMemHandlingEventRepository() cargo.HandlingEventRepository {
	return &handlingEventRepositoryInMem{
		events: make(map[cargo.TrackingID][]cargo.HandlingEvent),
	}
}
