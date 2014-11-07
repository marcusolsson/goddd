package inspection

import "github.com/marcusolsson/goddd/cargo"

type CargoEventHandler interface {
	CargoWasMisdirected(cargo.Cargo)
	CargoHasArrived(cargo.Cargo)
}

type CargoInspectionService interface {
	// InspectCargo inspects cargo and send relevant notifications to
	// interested parties, for example if a cargo has been misdirected, or
	// unloaded at the final destination.
	InspectCargo(trackingID cargo.TrackingID)
}

type cargoInspectionService struct {
	cargoRepository         cargo.Repository
	handlingEventRepository cargo.HandlingEventRepository
	cargoEventHandler       CargoEventHandler
}

// TODO: Should be transactional
func (s *cargoInspectionService) InspectCargo(trackingID cargo.TrackingID) {
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

// NewCargoInspectionService creates a inspection service with necessary
// dependencies.
func NewCargoInspectionService(cargoRepository cargo.Repository, handlingEventRepository cargo.HandlingEventRepository, eventHandler CargoEventHandler) CargoInspectionService {
	return &cargoInspectionService{cargoRepository, handlingEventRepository, eventHandler}
}
