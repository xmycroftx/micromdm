package device

import (
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// Service allows managing devices on the server
type Service interface {
	// Assigns a workflow to a device
	AssignWorkflow(workflowUUID, deviceUUID string) error
}

type service struct {
	devices Datastore
}

// NewService creates a new management service
func NewService(devices Datastore) Service {
	return &service{
		devices: devices,
	}
}

func (svc service) AssignWorkflow(workflowUUID string, deviceUDID string) error {
	device, err := svc.devices.GetDeviceByUDID(deviceUDID)
	if err != nil {
		return errors.Wrap(err, "assign workflow")
	}
	device.Workflow = workflowUUID
	err = svc.devices.Save("assign", device)
	if err != nil {
		return errors.Wrap(err, "assign workflow")
	}
	return nil
}

// ServiceHandler returns an http handler for the amanagement service
func ServiceHandler(ctx context.Context, svc Service) http.Handler {
	commonOptions := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(encodeError),
	}

	assignWorkflowHandler := httptransport.NewServer(
		ctx,
		makeAssignWorkflowEndpoint(svc),
		decodeAssignWorkflowRequest,
		encodeResponse,
		commonOptions...,
	)

	r := mux.NewRouter()
	r.Handle("/mdm/devices/{udid}/workflow", assignWorkflowHandler).Methods("POST")
	return r
}
