package tracking

import (
	"github.com/go-kit/kit/log"
)

type loggingService struct {
	Logger log.Logger
	Service
}

// NewLoggingService ...
func NewLoggingService(logger log.Logger, s Service) Service {
	return &loggingService{logger, s}
}

func (s *loggingService) Track(id string) (c Cargo, err error) {
	c, err = s.Service.Track(id)
	_ = s.Logger.Log(
		"method", "track",
		"err", err,
		"tracking_id", id,
	)
	return
}
