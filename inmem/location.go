package inmem

import shipping "github.com/marcusolsson/goddd"

type locationRepository struct {
	locations map[shipping.UNLocode]*shipping.Location
}

// NewLocationRepository returns a new instance of a in-memory location repository.
func NewLocationRepository() shipping.LocationRepository {
	r := &locationRepository{
		locations: make(map[shipping.UNLocode]*shipping.Location),
	}

	r.locations[shipping.SESTO] = shipping.Stockholm
	r.locations[shipping.AUMEL] = shipping.Melbourne
	r.locations[shipping.CNHKG] = shipping.Hongkong
	r.locations[shipping.JNTKO] = shipping.Tokyo
	r.locations[shipping.NLRTM] = shipping.Rotterdam
	r.locations[shipping.DEHAM] = shipping.Hamburg

	return r
}

func (r *locationRepository) Find(locode shipping.UNLocode) (*shipping.Location, error) {
	if l, ok := r.locations[locode]; ok {
		return l, nil
	}
	return nil, shipping.ErrUnknownLocation
}

func (r *locationRepository) FindAll() []*shipping.Location {
	l := make([]*shipping.Location, 0, len(r.locations))
	for _, val := range r.locations {
		l = append(l, val)
	}
	return l
}
