package booking

import (
	"time"

	"github.com/go-kit/kit/log"

	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
)

type loggingService struct {
	Logger log.Logger
	Service
}

// NewLoggingService ...
func NewLoggingService(logger log.Logger, s Service) Service {
	return &loggingService{logger, s}
}

func (s *loggingService) BookNewCargo(origin location.UNLocode, destination location.UNLocode, arrivalDeadline time.Time) (id cargo.TrackingID, err error) {
	id, err = s.Service.BookNewCargo(origin, destination, arrivalDeadline)
	_ = s.Logger.Log(
		"method", "book",
		"err", err,
		"origin", origin,
		"destination", destination,
		"arrival_deadline", arrivalDeadline,
		"tracking_id", id,
	)
	return
}

func (s *loggingService) ChangeDestination(id cargo.TrackingID, l location.UNLocode) (err error) {
	err = s.Service.ChangeDestination(id, l)
	_ = s.Logger.Log(
		"method", "change_destination",
		"err", err,
		"tracking_id", id,
		"destination", l,
	)
	return
}
