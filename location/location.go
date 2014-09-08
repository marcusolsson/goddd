package location

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
//
// It is uniquely identified by a UN Locode.
type Location struct {
	UNLocode UNLocode
	Name     string
}

var (
	// Special Location object that marks an unknown location.
	// What is the best way to handle this?
	UnknownLocation = Location{}
)

type LocationRepository interface {
	// Finds a location using given unlocode.
	Find(locode UNLocode) Location
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
