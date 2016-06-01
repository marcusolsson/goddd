package repository

import (
	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/voyage"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type mongoCargoRepository struct {
	db      string
	session *mgo.Session
}

func (r *mongoCargoRepository) Store(cargo *cargo.Cargo) (err error) {
	sess := r.session.Clone()
	defer sess.Close()

	c := sess.DB(r.db).C("cargo")

	index := mgo.Index{
		Key:        []string{"trackingid"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	if err = c.EnsureIndex(index); err != nil {
		return
	}

	_, err = c.Upsert(bson.M{"trackingid": cargo.TrackingID}, bson.M{"$set": cargo})

	return
}

func (r *mongoCargoRepository) Find(trackingID cargo.TrackingID) (*cargo.Cargo, error) {
	sess := r.session.Clone()
	defer sess.Close()

	c := sess.DB(r.db).C("cargo")

	var result cargo.Cargo
	if err := c.Find(bson.M{"trackingid": trackingID}).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			return nil, cargo.ErrUnknown
		}
		return nil, err
	}

	return &result, nil
}

func (r *mongoCargoRepository) FindAll() []*cargo.Cargo {
	sess := r.session.Clone()
	defer sess.Close()

	c := sess.DB(r.db).C("cargo")

	var result []*cargo.Cargo
	if err := c.Find(bson.M{}).All(&result); err != nil {
		return []*cargo.Cargo{}
	}

	return result
}

// NewMongoCargo returns a new instance of a MongoDB cargo repository.
func NewMongoCargo(db string, session *mgo.Session) cargo.Repository {
	return &mongoCargoRepository{
		db:      db,
		session: session,
	}
}

type mongoLocationRepository struct {
	db      string
	session *mgo.Session
}

func (r *mongoLocationRepository) Find(locode location.UNLocode) (location.Location, error) {
	sess := r.session.Clone()
	defer sess.Close()

	c := sess.DB(r.db).C("location")

	var result location.Location
	if err := c.Find(bson.M{"unlocode": locode}).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			return location.Location{}, location.ErrUnknown
		}
		return location.Location{}, err
	}

	return result, nil
}

func (r *mongoLocationRepository) FindAll() []location.Location {
	sess := r.session.Clone()
	defer sess.Close()

	c := sess.DB(r.db).C("location")

	var result []location.Location
	if err := c.Find(bson.M{}).All(&result); err != nil {
		return []location.Location{}
	}

	return result
}

func (r *mongoLocationRepository) store(l location.Location) (err error) {
	sess := r.session.Clone()
	defer sess.Close()

	c := sess.DB(r.db).C("location")

	index := mgo.Index{
		Key:        []string{"unlocode"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	if err = c.EnsureIndex(index); err != nil {
		return
	}

	_, err = c.Upsert(bson.M{"unlocode": l.UNLocode}, bson.M{"$set": l})

	return
}

// NewMongoLocation returns a new instance of a MongoDB location repository.
func NewMongoLocation(db string, session *mgo.Session) location.Repository {
	r := &mongoLocationRepository{
		db:      db,
		session: session,
	}

	initial := []location.Location{
		location.Stockholm,
		location.Melbourne,
		location.Hongkong,
		location.Tokyo,
		location.Rotterdam,
		location.Hamburg,
	}

	for _, l := range initial {
		r.store(l)
	}

	return r
}

type mongoVoyageRepository struct {
	db      string
	session *mgo.Session
}

func (r *mongoVoyageRepository) Find(voyageNumber voyage.Number) (*voyage.Voyage, error) {
	sess := r.session.Clone()
	defer sess.Close()

	c := sess.DB(r.db).C("voyage")

	var result voyage.Voyage
	if err := c.Find(bson.M{"number": voyageNumber}).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			return nil, voyage.ErrUnknown
		}
		return nil, err
	}

	return &result, nil
}

func (r *mongoVoyageRepository) store(v *voyage.Voyage) (err error) {
	sess := r.session.Clone()
	defer sess.Close()

	c := sess.DB(r.db).C("voyage")

	index := mgo.Index{
		Key:        []string{"number"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	if err = c.EnsureIndex(index); err != nil {
		return
	}

	_, err = c.Upsert(bson.M{"number": v.Number}, bson.M{"$set": v})

	return
}

// NewMongoVoyage returns a new instance of a MongoDB voyage repository.
func NewMongoVoyage(db string, session *mgo.Session) voyage.Repository {
	r := &mongoVoyageRepository{
		db:      db,
		session: session,
	}

	initial := []*voyage.Voyage{
		voyage.V100,
		voyage.V300,
		voyage.V400,
		voyage.V0100S,
		voyage.V0200T,
		voyage.V0300A,
		voyage.V0301S,
		voyage.V0400S,
	}

	for _, v := range initial {
		r.store(v)
	}

	return r
}

type mongoHandlingEventRepository struct {
	db      string
	session *mgo.Session
}

func (r *mongoHandlingEventRepository) Store(e cargo.HandlingEvent) {
	sess := r.session.Clone()
	defer sess.Close()

	c := sess.DB(r.db).C("handling_event")

	_ = c.Insert(e)
}

func (r *mongoHandlingEventRepository) QueryHandlingHistory(trackingID cargo.TrackingID) cargo.HandlingHistory {
	sess := r.session.Clone()
	defer sess.Close()

	c := sess.DB(r.db).C("handling_event")

	var result []cargo.HandlingEvent
	_ = c.Find(bson.M{"trackingid": trackingID}).All(&result)

	return cargo.HandlingHistory{HandlingEvents: result}
}

// NewMongoHandlingEvent returns a new instance of a MongoDB handling event repository.
func NewMongoHandlingEvent(db string, session *mgo.Session) cargo.HandlingEventRepository {
	return &mongoHandlingEventRepository{
		db:      db,
		session: session,
	}
}
