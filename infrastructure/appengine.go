package infrastructure

import (
	"github.com/marcusolsson/goddd/domain/location"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

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
