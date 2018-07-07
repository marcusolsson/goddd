package handling

import (
	"time"

	"github.com/go-kit/kit/metrics"

	shipping "github.com/marcusolsson/goddd"
)

type instrumentingService struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	next           Service
}

// NewInstrumentingService returns an instance of an instrumenting Service.
func NewInstrumentingService(counter metrics.Counter, latency metrics.Histogram, s Service) Service {
	return &instrumentingService{
		requestCount:   counter,
		requestLatency: latency,
		next:           s,
	}
}

func (s *instrumentingService) RegisterHandlingEvent(completed time.Time, id shipping.TrackingID, voyageNumber shipping.VoyageNumber,
	loc shipping.UNLocode, eventType shipping.HandlingEventType) error {

	defer func(begin time.Time) {
		s.requestCount.With("method", "register_incident").Add(1)
		s.requestLatency.With("method", "register_incident").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.RegisterHandlingEvent(completed, id, voyageNumber, loc, eventType)
}
