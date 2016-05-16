package management

import (
	"encoding/json"
	"net/http"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/micromdm/micromdm/device"
)

type listDevicesRequest struct{}

type listDevicesResponse struct {
	devices []device.Device
	Err     error `json:"error,omitempty"`
}

func (r listDevicesResponse) error() error { return r.Err }

func (r listDevicesResponse) encodeList(w http.ResponseWriter) error {
	jsn, err := json.MarshalIndent(r.devices, "", "  ")
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(jsn)
	return nil
}

func makeListDevicesEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		ds, err := svc.Devices()
		return listDevicesResponse{Err: err, devices: ds}, nil
	}
}
