package tracking

import (
	"github.com/go-kit/kit/endpoint"
	"github.com/marcusolsson/goddd/cargo"
	"golang.org/x/net/context"
)

type findCargoRequest struct {
	ID string
}

type findCargoResponse struct {
	Cargo *Cargo `json:"cargo,omitempty"`
	Err   string `json:"error,omitempty"`
}

func makeFindCargoEndpoint(ts Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(findCargoRequest)

		c, err := ts.Track(req.ID)
		if err != nil {
			if err == cargo.ErrUnknown {
				return findCargoResponse{Err: err.Error()}, nil
			}
			return nil, err
		}

		return findCargoResponse{Cargo: &c}, nil
	}
}
