package handling

import (
	"time"

	"github.com/go-kit/kit/log"

	shipping "github.com/marcusolsson/goddd"
)

type loggingService struct {
	logger log.Logger
	next   Service
}

// NewLoggingService returns a new instance of a logging Service.
func NewLoggingService(logger log.Logger, s Service) Service {
	return &loggingService{logger, s}
}

func (s *loggingService) RegisterHandlingEvent(completed time.Time, id shipping.TrackingID, voyageNumber shipping.VoyageNumber,
	unLocode shipping.UNLocode, eventType shipping.HandlingEventType) (err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "register_incident",
			"tracking_id", id,
			"location", unLocode,
			"voyage", voyageNumber,
			"event_type", eventType,
			"completion_time", completed,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.next.RegisterHandlingEvent(completed, id, voyageNumber, unLocode, eventType)
}
