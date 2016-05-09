package device

import (
	"errors"

	"github.com/go-kit/kit/endpoint"
	"golang.org/x/net/context"
)

var (
	// ErrEmptyRequest is returned if the request body is empty
	ErrEmptyRequest = errors.New("request must contain a profile identifier")
	errBadRouting   = errors.New("inconsistent mapping between route and handler (programmer error)")
)

func makeAssignWorkflowEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(assignWorkflowRequest)
		if req.DeviceUDID == "" || req.WorkflowUUID == "" {
			return assignWorkflowResponse{Err: ErrEmptyRequest}, nil
		}
		err := svc.AssignWorkflow(req.WorkflowUUID, req.DeviceUDID)
		if err != nil {
			return assignWorkflowResponse{Err: err}, nil
		}
		return assignWorkflowResponse{}, nil
	}
}
