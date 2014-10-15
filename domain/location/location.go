package location

import (
	"errors"

	"github.com/marcusolsson/goddd/domain/shared"
)

// UNLocode is the United Nations location code that uniquely identifies a
// particular location.
//
// http://www.unece.org/cefact/locode/
// http://www.unece.org/cefact/locode/DocColumnDescription.htm#LOCODE
type UNLocode string

// Location is a location is our model is stops on a journey, such as cargo
// origin or destination, or carrier movement endpoints.
type Location struct {
	UNLocode UNLocode
	Name     string
}

func (l Location) SameIdentity(e shared.Entity) bool {
	return l.UNLocode == e.(Location).UNLocode
}

// ErrUnknownLocation is used when a location could not be found.
var ErrUnknownLocation = errors.New("unknown location")

// LocationRepository provides access a location store.
type LocationRepository interface {
	Find(locode UNLocode) (Location, error)
	FindAll() []Location
}
