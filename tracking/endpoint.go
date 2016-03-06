package tracking

import (
	"github.com/go-kit/kit/endpoint"
	"golang.org/x/net/context"
)

type findCargoRequest struct {
	ID string
}

type findCargoResponse struct {
	Cargo *Cargo `json:"cargo,omitempty"`
	Err   error  `json:"error,omitempty"`
}

func (r findCargoResponse) error() error { return r.Err }

func makeFindCargoEndpoint(ts Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(findCargoRequest)
		c, err := ts.Track(req.ID)
		return findCargoResponse{Cargo: &c, Err: err}, nil
	}
}
