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

type showDeviceRequest struct {
	UUID string
}

type showDeviceResponse struct {
	*device.Device
	Err error `json:"error,omitempty"`
}

func (r showDeviceResponse) error() error { return r.Err }

func makeShowDeviceEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(showDeviceRequest)
		dev, err := svc.Device(req.UUID)
		if err != nil {
			return showDeviceResponse{Err: err}, nil
		}
		return showDeviceResponse{Device: dev}, nil
	}
}

type updateDeviceRequest struct {
	DeviceUUID string  `json:"-"`
	Workflow   *string `json:"workflow_uuid,omitempty" db:"workflow_uuid,omitempty"`
}

type updateDeviceResponse struct {
	Err error `json:"error,omitempty"`
}

func (r updateDeviceResponse) error() error { return r.Err }
func (r updateDeviceResponse) status() int  { return http.StatusNoContent }

func makeUpdateDeviceEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateDeviceRequest)
		var err error
		if req.Workflow != nil {
			err = svc.AssignWorkflow(req.DeviceUUID, *req.Workflow)
		}
		return updateDeviceResponse{Err: err}, nil
	}
}
