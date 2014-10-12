package location

import (
	"errors"

	"github.com/marcusolsson/goddd/domain/shared"
)

var (
	SESTO UNLocode = "SESTO"
	AUMEL UNLocode = "AUMEL"
	CNHKG UNLocode = "CNHKG"
	USNYC UNLocode = "USNYC"
	USCHI UNLocode = "USCHI"
	JNTKO UNLocode = "JNTKO"
	DEHAM UNLocode = "DEHAM"
)

var (
	Stockholm = Location{SESTO, "Stockholm"}
	Melbourne = Location{AUMEL, "Melbourne"}
	Hongkong  = Location{CNHKG, "Hongkong"}
	NewYork   = Location{USNYC, "New York"}
	Chicago   = Location{USCHI, "Chicago"}
	Tokyo     = Location{JNTKO, "Tokyo"}
	Hamburg   = Location{DEHAM, "Hamburg"}
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

var ErrUnknownLocation = errors.New("unknown location")

type LocationRepository interface {
	Find(locode UNLocode) (Location, error)
	FindAll() []Location
}
