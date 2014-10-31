package voyage

import "github.com/marcusolsson/goddd/domain/location"

var (
	V100 = New("V100", Schedule{
		[]CarrierMovement{
			CarrierMovement{DepartureLocation: location.Hongkong, ArrivalLocation: location.Tokyo},
			CarrierMovement{DepartureLocation: location.Tokyo, ArrivalLocation: location.NewYork},
		},
	})

	V300 = New("V300", Schedule{
		[]CarrierMovement{
			CarrierMovement{DepartureLocation: location.Tokyo, ArrivalLocation: location.Rotterdam},
			CarrierMovement{DepartureLocation: location.Rotterdam, ArrivalLocation: location.Hamburg},
			CarrierMovement{DepartureLocation: location.Hamburg, ArrivalLocation: location.Melbourne},
			CarrierMovement{DepartureLocation: location.Melbourne, ArrivalLocation: location.Tokyo},
		},
	})

	V400 = New("V400", Schedule{
		[]CarrierMovement{
			CarrierMovement{DepartureLocation: location.Hamburg, ArrivalLocation: location.Stockholm},
			CarrierMovement{DepartureLocation: location.Stockholm, ArrivalLocation: location.Helsinki},
			CarrierMovement{DepartureLocation: location.Helsinki, ArrivalLocation: location.Hamburg},
		},
	})

	// These voyages are hard-coded into the current pathfinder. Make sure
	// they exist.
	V0100S = New("0100S", Schedule{[]CarrierMovement{}})
	V0200T = New("0200T", Schedule{[]CarrierMovement{}})
	V0300A = New("0300A", Schedule{[]CarrierMovement{}})
	V0301S = New("0301S", Schedule{[]CarrierMovement{}})
	V0400S = New("0400S", Schedule{[]CarrierMovement{}})
)
