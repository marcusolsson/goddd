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
)
