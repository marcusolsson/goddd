package location

import "github.com/marcusolsson/goddd/domain/shared"

var (
	Stockholm = Location{UNLocode: "SESTO", Name: "Stockholm"}
	Melbourne = Location{UNLocode: "AUMEL", Name: "Melbourne"}
	Hongkong  = Location{UNLocode: "CNHKG", Name: "Hongkong"}
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

type locationRepository struct {
	locations map[UNLocode]Location
}

func (r *locationRepository) Find(locode UNLocode) Location {
	if l, ok := r.locations[locode]; ok {
		return l
	}

	return UnknownLocation
}

func (r *locationRepository) FindAll() []Location {
	l := make([]Location, 0, len(r.locations))
	for _, val := range r.locations {
		l = append(l, val)
	}
	return l
}

func NewLocationRepository() LocationRepository {
	r := &locationRepository{
		locations: make(map[UNLocode]Location),
	}

	r.locations[Stockholm.UNLocode] = Stockholm
	r.locations[Melbourne.UNLocode] = Melbourne
	r.locations[Hongkong.UNLocode] = Hongkong

	return r
}
