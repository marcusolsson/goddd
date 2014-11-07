package inspection

import "github.com/marcusolsson/goddd/cargo"

type EventHandler interface {
	CargoWasMisdirected(cargo.Cargo)
	CargoHasArrived(cargo.Cargo)
}

type Service interface {
	// InspectCargo inspects cargo and send relevant notifications to
	// interested parties, for example if a cargo has been misdirected, or
	// unloaded at the final destination.
	InspectCargo(trackingID cargo.TrackingID)
}

type service struct {
	cargoRepository         cargo.Repository
	handlingEventRepository cargo.HandlingEventRepository
	cargoEventHandler       EventHandler
}

// TODO: Should be transactional
func (s *service) InspectCargo(trackingID cargo.TrackingID) {
	c, err := s.cargoRepository.Find(trackingID)

	if err != nil {
		return
	}

	h := s.handlingEventRepository.QueryHandlingHistory(trackingID)

	c.DeriveDeliveryProgress(h)

	if c.Delivery.IsMisdirected {
		s.cargoEventHandler.CargoWasMisdirected(c)
	}

	if c.Delivery.IsUnloadedAtDestination {
		s.cargoEventHandler.CargoHasArrived(c)
	}

	s.cargoRepository.Store(c)
}

// NewService creates a inspection service with necessary
// dependencies.
func NewService(cargoRepository cargo.Repository, handlingEventRepository cargo.HandlingEventRepository, eventHandler EventHandler) Service {
	return &service{cargoRepository, handlingEventRepository, eventHandler}
}
