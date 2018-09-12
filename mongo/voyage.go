package mongo

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"

	shipping "github.com/marcusolsson/goddd"
)

const voyageCollection = "voyage"

type voyageRepository struct {
	db      string
	session *mgo.Session
}

// NewVoyageRepository returns a new instance of a MongoDB voyage repository.
func NewVoyageRepository(db string, session *mgo.Session) (shipping.VoyageRepository, error) {
	r := &voyageRepository{
		db:      db,
		session: session,
	}

	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C(voyageCollection)

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

func (r *voyageRepository) Find(voyageNumber shipping.VoyageNumber) (*shipping.Voyage, error) {
	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C(voyageCollection)

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

	c := sess.DB(r.db).C(voyageCollection)

	_, err := c.Upsert(bson.M{"number": v.VoyageNumber}, bson.M{"$set": v})

	return err
}
