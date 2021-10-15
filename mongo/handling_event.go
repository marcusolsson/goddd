package mongo

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"

	shipping "github.com/marcusolsson/goddd"
)

const handlingEventCollection = "handling_event"

type handlingEventRepository struct {
	db      string
	session *mgo.Session
}

// NewHandlingEventRepository returns a new instance of a MongoDB handling event repository.
func NewHandlingEventRepository(db string, session *mgo.Session) shipping.HandlingEventRepository {
	return &handlingEventRepository{
		db:      db,
		session: session,
	}
}

func (r *handlingEventRepository) Store(e shipping.HandlingEvent) {
	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C(handlingEventCollection)

	_ = c.Insert(e)
}

func (r *handlingEventRepository) QueryHandlingHistory(id shipping.TrackingID) shipping.HandlingHistory {
	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C(handlingEventCollection)

	var result []shipping.HandlingEvent
	_ = c.Find(bson.M{"trackingid": id}).All(&result)

	return shipping.HandlingHistory{HandlingEvents: result}
}
