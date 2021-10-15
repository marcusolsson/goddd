package tracking

import (
	"time"

	"github.com/go-kit/kit/log"
)

type loggingMiddleware struct {
	logger log.Logger
	next   Service
}

// NewLoggingMiddleware returns a new instance of a logging middleware.
func NewLoggingMiddleware(logger log.Logger, next Service) Service {
	return &loggingMiddleware{logger, next}
}

func (mw *loggingMiddleware) Track(id string) (c Cargo, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "track", "tracking_id", id, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.Track(id)
}
