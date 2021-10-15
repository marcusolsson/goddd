package mongo

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"

	shipping "github.com/marcusolsson/goddd"
)

const cargoCollection = "cargo"

type cargoRepository struct {
	db      string
	session *mgo.Session
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

	c := sess.DB(r.db).C(cargoCollection)

	if err := c.EnsureIndex(index); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *cargoRepository) Store(cargo *shipping.Cargo) error {
	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C(cargoCollection)

	_, err := c.Upsert(bson.M{"trackingid": cargo.TrackingID}, bson.M{"$set": cargo})

	return err
}

func (r *cargoRepository) Find(id shipping.TrackingID) (*shipping.Cargo, error) {
	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C(cargoCollection)

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

	c := sess.DB(r.db).C(cargoCollection)

	var result []*shipping.Cargo
	if err := c.Find(bson.M{}).All(&result); err != nil {
		return []*shipping.Cargo{}
	}

	return result
}
