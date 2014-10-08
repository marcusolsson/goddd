package application

import (
	"container/list"

	"github.com/marcusolsson/goddd/domain/cargo"
)

type CargoInspectionService interface {
	InspectCargo(trackingId cargo.TrackingId)
}

type cargoInspectionService struct {
	cargoRepository cargo.CargoRepository
	eventHandler    EventHandler
}

func (s *cargoInspectionService) InspectCargo(trackingId cargo.TrackingId) {
	c, err := s.cargoRepository.Find(trackingId)

	if err != nil {
		return
	}

	c.DeriveDeliveryProgress(cargo.HandlingHistory{list.New()})

	s.eventHandler.CargoWasMisdirected(c)
}
