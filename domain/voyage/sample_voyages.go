package voyage

import "github.com/marcusolsson/goddd/domain/location"

var (
	V100 = New("V100", Schedule{
		[]CarrierMovement{
			CarrierMovement{DepartureLocation: location.Hongkong, ArrivalLocation: location.Tokyo},
			CarrierMovement{DepartureLocation: location.Tokyo, ArrivalLocation: location.NewYork},
		},
	})
)
