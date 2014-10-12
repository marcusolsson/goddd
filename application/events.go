package application

import "github.com/marcusolsson/goddd/domain/cargo"

type HandlingEventHandler interface {
	CargoWasHandled(cargo.HandlingEvent)
}

type CargoEventHandler interface {
	CargoWasMisdirected(cargo.Cargo)
	CargoHasArrived(cargo.Cargo)
}
