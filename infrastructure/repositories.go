package infrastructure

import (
	"github.com/marcusolsson/goddd/domain/cargo"
	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/voyage"
)

type cargoRepositoryInMem struct {
	cargos map[cargo.TrackingId]cargo.Cargo
}

func (r *cargoRepositoryInMem) Store(c cargo.Cargo) error {
	r.cargos[c.TrackingId] = c

	return nil
}

func (r *cargoRepositoryInMem) Find(trackingId cargo.TrackingId) (cargo.Cargo, error) {

	if val, ok := r.cargos[trackingId]; ok {
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

func NewInMemCargoRepository() cargo.CargoRepository {
	return &cargoRepositoryInMem{
		cargos: make(map[cargo.TrackingId]cargo.Cargo),
	}
}

type locationRepositoryInMem struct {
	locations map[location.UNLocode]location.Location
}

func (r *locationRepositoryInMem) Find(locode location.UNLocode) location.Location {
	if l, ok := r.locations[locode]; ok {
		return l
	}

	return location.UnknownLocation
}

func (r *locationRepositoryInMem) FindAll() []location.Location {
	l := make([]location.Location, 0, len(r.locations))
	for _, val := range r.locations {
		l = append(l, val)
	}
	return l
}

func NewInMemLocationRepository() location.LocationRepository {
	r := &locationRepositoryInMem{
		locations: make(map[location.UNLocode]location.Location),
	}

	r.locations[location.Stockholm.UNLocode] = location.Stockholm
	r.locations[location.Melbourne.UNLocode] = location.Melbourne
	r.locations[location.Hongkong.UNLocode] = location.Hongkong

	return r
}

type voyageRepositoryInMem struct {
	voyages map[voyage.VoyageNumber]voyage.Voyage
}

func (r *voyageRepositoryInMem) Find(voyageNumber voyage.VoyageNumber) voyage.Voyage {
	if v, ok := r.voyages[voyageNumber]; ok {
		return v
	}

	return voyage.Voyage{}
}

func NewInMemVoyageRepository() voyage.VoyageRepository {
	r := &voyageRepositoryInMem{
		voyages: make(map[voyage.VoyageNumber]voyage.Voyage),
	}

	r.voyages[voyage.MelbourneToStockholm.VoyageNumber] = *voyage.MelbourneToStockholm

	return r
}
