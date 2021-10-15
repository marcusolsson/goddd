package mongo

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"

	shipping "github.com/marcusolsson/goddd"
)

const locationCollection = "location"

type locationRepository struct {
	db      string
	session *mgo.Session
}

// NewLocationRepository returns a new instance of a MongoDB location repository.
func NewLocationRepository(db string, session *mgo.Session) (shipping.LocationRepository, error) {
	r := &locationRepository{
		db:      db,
		session: session,
	}

	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C(locationCollection)

	index := mgo.Index{
		Key:        []string{"unlocode"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	if err := c.EnsureIndex(index); err != nil {
		return nil, err
	}

	initial := []*shipping.Location{
		shipping.Stockholm,
		shipping.Melbourne,
		shipping.Hongkong,
		shipping.Tokyo,
		shipping.Rotterdam,
		shipping.Hamburg,
	}

	for _, l := range initial {
		r.store(l)
	}

	return r, nil
}

func (r *locationRepository) Find(locode shipping.UNLocode) (*shipping.Location, error) {
	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C(locationCollection)

	var result shipping.Location
	if err := c.Find(bson.M{"unlocode": locode}).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			return nil, shipping.ErrUnknownLocation
		}
		return nil, err
	}

	return &result, nil
}

func (r *locationRepository) FindAll() []*shipping.Location {
	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C(locationCollection)

	var result []*shipping.Location
	if err := c.Find(bson.M{}).All(&result); err != nil {
		return []*shipping.Location{}
	}

	return result
}

func (r *locationRepository) store(l *shipping.Location) error {
	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C(locationCollection)

	_, err := c.Upsert(bson.M{"unlocode": l.UNLocode}, bson.M{"$set": l})

	return err
}
