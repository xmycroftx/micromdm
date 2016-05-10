package management

import (
	"github.com/go-kit/kit/endpoint"
	"golang.org/x/net/context"
)

type fetchDEPDevicesRequest struct{}

type fetchDEPDevicesResponse struct {
	Err error `json:"error,omitempty"`
}

func (r fetchDEPDevicesResponse) error() error { return r.Err }

func makeFetchDevicesEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		err := svc.FetchDEPDevices()
		return fetchDEPDevicesResponse{Err: err}, nil
	}
}
