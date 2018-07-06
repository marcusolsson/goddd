package mongo

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	shipping "github.com/marcusolsson/goddd"
)

type cargoRepository struct {
	db      string
	session *mgo.Session
}

func (r *cargoRepository) Store(cargo *shipping.Cargo) error {
	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C("cargo")

	_, err := c.Upsert(bson.M{"trackingid": cargo.TrackingID}, bson.M{"$set": cargo})

	return err
}

func (r *cargoRepository) Find(id shipping.TrackingID) (*shipping.Cargo, error) {
	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C("cargo")

	var result shipping.Cargo
	if err := c.Find(bson.M{"trackingid": id}).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			return nil, shipping.ErrUnknownCargo
		}
		return nil, err
	}

	return &result, nil
}

func (r *cargoRepository) FindAll() []*shipping.Cargo {
	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C("cargo")

	var result []*shipping.Cargo
	if err := c.Find(bson.M{}).All(&result); err != nil {
		return []*shipping.Cargo{}
	}

	return result
}

// NewCargoRepository returns a new instance of a MongoDB cargo repository.
func NewCargoRepository(db string, session *mgo.Session) (shipping.CargoRepository, error) {
	r := &cargoRepository{
		db:      db,
		session: session,
	}

	index := mgo.Index{
		Key:        []string{"trackingid"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C("cargo")

	if err := c.EnsureIndex(index); err != nil {
		return nil, err
	}

	return r, nil
}

type locationRepository struct {
	db      string
	session *mgo.Session
}

func (r *locationRepository) Find(locode shipping.UNLocode) (*shipping.Location, error) {
	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C("location")

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

	c := sess.DB(r.db).C("location")

	var result []*shipping.Location
	if err := c.Find(bson.M{}).All(&result); err != nil {
		return []*shipping.Location{}
	}

	return result
}

func (r *locationRepository) store(l *shipping.Location) error {
	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C("location")

	_, err := c.Upsert(bson.M{"unlocode": l.UNLocode}, bson.M{"$set": l})

	return err
}

// NewLocationRepository returns a new instance of a MongoDB location repository.
func NewLocationRepository(db string, session *mgo.Session) (shipping.LocationRepository, error) {
	r := &locationRepository{
		db:      db,
		session: session,
	}

	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C("location")

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

type voyageRepository struct {
	db      string
	session *mgo.Session
}

func (r *voyageRepository) Find(voyageNumber shipping.VoyageNumber) (*shipping.Voyage, error) {
	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C("voyage")

	var result shipping.Voyage
	if err := c.Find(bson.M{"number": voyageNumber}).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			return nil, shipping.ErrUnknownVoyage
		}
		return nil, err
	}

	return &result, nil
}

func (r *voyageRepository) store(v *shipping.Voyage) error {
	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C("voyage")

	_, err := c.Upsert(bson.M{"number": v.VoyageNumber}, bson.M{"$set": v})

	return err
}

// NewVoyageRepository returns a new instance of a MongoDB voyage repository.
func NewVoyageRepository(db string, session *mgo.Session) (shipping.VoyageRepository, error) {
	r := &voyageRepository{
		db:      db,
		session: session,
	}

	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C("voyage")

	index := mgo.Index{
		Key:        []string{"number"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	if err := c.EnsureIndex(index); err != nil {
		return nil, err
	}

	initial := []*shipping.Voyage{
		shipping.V100,
		shipping.V300,
		shipping.V400,
		shipping.V0100S,
		shipping.V0200T,
		shipping.V0300A,
		shipping.V0301S,
		shipping.V0400S,
	}

	for _, v := range initial {
		r.store(v)
	}

	return r, nil
}

type handlingEventRepository struct {
	db      string
	session *mgo.Session
}

func (r *handlingEventRepository) Store(e shipping.HandlingEvent) {
	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C("handling_event")

	_ = c.Insert(e)
}

func (r *handlingEventRepository) QueryHandlingHistory(id shipping.TrackingID) shipping.HandlingHistory {
	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C("handling_event")

	var result []shipping.HandlingEvent
	_ = c.Find(bson.M{"trackingid": id}).All(&result)

	return shipping.HandlingHistory{HandlingEvents: result}
}

// NewHandlingEventRepository returns a new instance of a MongoDB handling event repository.
func NewHandlingEventRepository(db string, session *mgo.Session) shipping.HandlingEventRepository {
	return &handlingEventRepository{
		db:      db,
		session: session,
	}
}
