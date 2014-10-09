package location

import "github.com/marcusolsson/goddd/domain/shared"

var (
	Stockholm = Location{UNLocode: "SESTO", Name: "Stockholm"}
	Melbourne = Location{UNLocode: "AUMEL", Name: "Melbourne"}
	Hongkong  = Location{UNLocode: "CNHKG", Name: "Hongkong"}
	NewYork   = Location{UNLocode: "USNYC", Name: "New York"}
)

type UNLocode string

type Location struct {
	UNLocode UNLocode
	Name     string
}

func (l Location) SameIdentity(e shared.Entity) bool {
	return l.UNLocode == e.(Location).UNLocode
}

var UnknownLocation = Location{}

type LocationRepository interface {
	Find(locode UNLocode) Location
	FindAll() []Location
}
