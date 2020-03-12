package handling

import (
	"time"

	"github.com/go-kit/kit/metrics"

	shipping "github.com/marcusolsson/goddd"
)

type instrumentingMiddleware struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	next           Service
}

// NewInstrumentingMiddleware returns an instance of an instrumenting middleware.
func NewInstrumentingMiddleware(counter metrics.Counter, latency metrics.Histogram, next Service) Service {
	return &instrumentingMiddleware{
		requestCount:   counter,
		requestLatency: latency,
		next:           next,
	}
}

func (mw *instrumentingMiddleware) RegisterHandlingEvent(completed time.Time, id shipping.TrackingID, voyageNumber shipping.VoyageNumber,
	loc shipping.UNLocode, eventType shipping.HandlingEventType) error {

	defer func(begin time.Time) {
		mw.requestCount.With("method", "register_incident").Add(1)
		mw.requestLatency.With("method", "register_incident").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.RegisterHandlingEvent(completed, id, voyageNumber, loc, eventType)
}
