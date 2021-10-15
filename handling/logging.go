package handling

import (
	"time"

	"github.com/go-kit/kit/log"

	shipping "github.com/marcusolsson/goddd"
)

type loggingMiddleware struct {
	logger log.Logger
	next   Service
}

// NewLoggingMiddleware returns a new instance of a logging middleware.
func NewLoggingMiddleware(logger log.Logger, next Service) Service {
	return &loggingMiddleware{logger, next}
}

func (mw *loggingMiddleware) RegisterHandlingEvent(completed time.Time, id shipping.TrackingID, voyageNumber shipping.VoyageNumber,
	unLocode shipping.UNLocode, eventType shipping.HandlingEventType) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
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
	return mw.next.RegisterHandlingEvent(completed, id, voyageNumber, unLocode, eventType)
}
