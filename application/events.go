package application

import "github.com/marcusolsson/goddd/domain/cargo"

type EventHandler interface {
	CargoWasHandled(cargo.HandlingEvent)
	CargoWasMisdirected(cargo.Cargo)
	CargoHasArrived(cargo.Cargo)
}
