package inmem

import shipping "github.com/marcusolsson/goddd"

type voyageRepository struct {
	voyages map[shipping.VoyageNumber]*shipping.Voyage
}

// NewVoyageRepository returns a new instance of a in-memory voyage repository.
func NewVoyageRepository() shipping.VoyageRepository {
	r := &voyageRepository{
		voyages: make(map[shipping.VoyageNumber]*shipping.Voyage),
	}

	r.voyages[shipping.V100.VoyageNumber] = shipping.V100
	r.voyages[shipping.V300.VoyageNumber] = shipping.V300
	r.voyages[shipping.V400.VoyageNumber] = shipping.V400

	r.voyages[shipping.V0100S.VoyageNumber] = shipping.V0100S
	r.voyages[shipping.V0200T.VoyageNumber] = shipping.V0200T
	r.voyages[shipping.V0300A.VoyageNumber] = shipping.V0300A
	r.voyages[shipping.V0301S.VoyageNumber] = shipping.V0301S
	r.voyages[shipping.V0400S.VoyageNumber] = shipping.V0400S

	return r
}

func (r *voyageRepository) Find(voyageNumber shipping.VoyageNumber) (*shipping.Voyage, error) {
	if v, ok := r.voyages[voyageNumber]; ok {
		return v, nil
	}

	return nil, shipping.ErrUnknownVoyage
}
