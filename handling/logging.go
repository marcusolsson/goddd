package handling

import (
	"time"

	"github.com/go-kit/kit/log"

	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/voyage"
)

type loggingService struct {
	Logger log.Logger
	Service
}

// NewLoggingService ...
func NewLoggingService(logger log.Logger, s Service) Service {
	return &loggingService{logger, s}
}

func (s *loggingService) RegisterHandlingEvent(completionTime time.Time, trackingID cargo.TrackingID, voyageNumber voyage.Number,
	unLocode location.UNLocode, eventType cargo.HandlingEventType) (err error) {
	err = s.Service.RegisterHandlingEvent(completionTime, trackingID, voyageNumber, unLocode, eventType)
	_ = s.Logger.Log(
		"method", "register_incident",
		"err", err,
		"event_type", eventType,
		"location", unLocode,
		"voyage_number", voyageNumber,
		"tracking_id", trackingID,
	)
	return
}
