package infrastructure

import (
	"github.com/marcusolsson/goddd/domain/cargo"
	"github.com/marcusolsson/goddd/domain/location"
	"github.com/marcusolsson/goddd/domain/voyage"

	"appengine"
	"appengine/datastore"
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

func NewInMemLocationRepository() location.LocationRepository {
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

type locationRepositoryDatastore struct {
	C appengine.Context
}

func (r *locationRepositoryDatastore) Store(l location.Location) {
	key := datastore.NewKey(r.C, "location", string(l.UNLocode), 0, nil)
	key, err := datastore.Put(r.C, key, &l)
	if err != nil {
		panic(err)
		return
	}
}

func (r *locationRepositoryDatastore) Find(locode location.UNLocode) (location.Location, error) {
	var l location.Location
	key := datastore.NewKey(r.C, "location", string(locode), 0, nil)
	if err := datastore.Get(r.C, key, &l); err != nil {
		return location.Location{}, err
	}

	return l, nil
}

func (r *locationRepositoryDatastore) FindAll() []location.Location {
	q := datastore.NewQuery("location")
	var locations []location.Location
	_, err := q.GetAll(r.C, &locations)

	if err != nil {
		return []location.Location{}
	}

	return locations
}

func NewDatastoreLocationRepository(context appengine.Context) *locationRepositoryDatastore {
	r := &locationRepositoryDatastore{C: context}
	r.Store(location.Stockholm)
	r.Store(location.Melbourne)
	r.Store(location.Hamburg)

	return r
}

type voyageRepositoryInMem struct {
	voyages map[voyage.VoyageNumber]voyage.Voyage
}

func (r *voyageRepositoryInMem) Find(voyageNumber voyage.VoyageNumber) (voyage.Voyage, error) {
	if v, ok := r.voyages[voyageNumber]; ok {
		return v, nil
	}

	return voyage.Voyage{}, voyage.ErrUnknownVoyage
}

func NewInMemVoyageRepository() voyage.VoyageRepository {
	r := &voyageRepositoryInMem{
		voyages: make(map[voyage.VoyageNumber]voyage.Voyage),
	}

	r.voyages[voyage.V100.VoyageNumber] = *voyage.V100
	r.voyages[voyage.V300.VoyageNumber] = *voyage.V300
	r.voyages[voyage.V400.VoyageNumber] = *voyage.V400

	return r
}

type handlingEventRepositoryInMem struct {
	events map[cargo.TrackingId][]cargo.HandlingEvent
}

func (r *handlingEventRepositoryInMem) Store(e cargo.HandlingEvent) {
	// Make array if it's the first event with this tracking ID.
	if _, ok := r.events[e.TrackingId]; !ok {
		r.events[e.TrackingId] = make([]cargo.HandlingEvent, 0)
	}
	r.events[e.TrackingId] = append(r.events[e.TrackingId], e)
}

func (r *handlingEventRepositoryInMem) QueryHandlingHistory(trackingId cargo.TrackingId) cargo.HandlingHistory {
	return cargo.HandlingHistory{r.events[trackingId]}
}

func NewInMemHandlingEventRepository() cargo.HandlingEventRepository {
	return &handlingEventRepositoryInMem{
		events: make(map[cargo.TrackingId][]cargo.HandlingEvent),
	}
}
