package location

import "github.com/marcusolsson/goddd/domain/shared"

// Some samples locations
var (
	Stockholm = Location{
		UNLocode: "SESTO",
		Name:     "Stockholm",
	}
	Melbourne = Location{
		UNLocode: "AUMEL",
		Name:     "Melbourne",
	}
	Hongkong = Location{
		UNLocode: "CNHKG",
		Name:     "Hongkong",
	}
)

type UNLocode string

// Location is our model is stops on a journey, such as cargo origin
// or destination, or carrier movement endpoints.
type Location struct {
	UNLocode UNLocode
	Name     string
}

func (l Location) SameIdentity(e shared.Entity) bool {
	return l.UNLocode == e.(Location).UNLocode
}

var (
	// Special Location object that marks an unknown location.
	// What is the best way to handle this?
	UnknownLocation = Location{}
)

type LocationRepository interface {
	// Finds a location using given unlocode.
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

	// Populate with test data
	r.locations[Stockholm.UNLocode] = Stockholm
	r.locations[Melbourne.UNLocode] = Melbourne
	r.locations[Hongkong.UNLocode] = Hongkong

	return r
}
