package management

import (
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/micromdm/micromdm/workflow"
	"golang.org/x/net/context"
)

type addWorkflowRequest struct {
	*workflow.Workflow
}

type addWorkflowResponse struct {
	*workflow.Workflow
	Err error `json:"error,omitempty"`
}

func (r addWorkflowResponse) status() int { return http.StatusCreated }

func (r addWorkflowResponse) error() error { return r.Err }

func makeAddWorkflowEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(addWorkflowRequest)
		wf, err := svc.AddWorkflow(req.Workflow)
		return addWorkflowResponse{Err: err, Workflow: wf}, nil
	}
}

type listWorkflowsRequest struct{}

type listWorkflowsResponse struct {
	workflows []workflow.Workflow
	Err       error `json:"error,omitempty"`
}

func (r listWorkflowsResponse) error() error { return r.Err }

func (r listWorkflowsResponse) encodeList(w http.ResponseWriter) error {
	jsn, err := json.MarshalIndent(r.workflows, "", "  ")
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(jsn)
	return nil
}

func makeListWorkflowsEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		workflows, err := svc.Workflows()
		return listWorkflowsResponse{Err: err, workflows: workflows}, nil
	}
}
