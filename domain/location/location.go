package location

import (
	"errors"

	"github.com/marcusolsson/goddd/domain/shared"
)

type UNLocode string

type Location struct {
	UNLocode UNLocode
	Name     string
}

func (l Location) SameIdentity(e shared.Entity) bool {
	return l.UNLocode == e.(Location).UNLocode
}

var ErrUnknownLocation = errors.New("unknown location")

type LocationRepository interface {
	Find(locode UNLocode) (Location, error)
	FindAll() []Location
}
