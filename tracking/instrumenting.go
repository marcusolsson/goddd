package tracking

import (
	"time"

	"github.com/go-kit/kit/metrics"
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

func (mw *instrumentingMiddleware) Track(id string) (Cargo, error) {
	defer func(begin time.Time) {
		mw.requestCount.With("method", "track").Add(1)
		mw.requestLatency.With("method", "track").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.Track(id)
}
