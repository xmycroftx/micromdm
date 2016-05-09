package workflow

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"golang.org/x/net/context"
)

var (
	// ErrEmptyRequest is returned if the request body is empty
	ErrEmptyRequest = errors.New("request must contain a name")
	errBadRouting   = errors.New("inconsistent mapping between route and handler (programmer error)")
)

// NewWorkflowRequest in an HTTP request for a new workflow
type NewWorkflowRequest struct {
	*Workflow
	// Name string `json:"name"`
}

func decodeNewWorkflowRequest(r *http.Request) (interface{}, error) {
	var request NewWorkflowRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	return request, err
}

// NewWorkflowResponse is a command reponse
type NewWorkflowResponse struct {
	*Workflow
	Err error `json:"error,omitempty"`
}

// for API statuses other than 200OK
type statuser interface {
	status() int
}

func (r NewWorkflowResponse) status() int { return 201 }

func (r NewWorkflowResponse) error() error { return r.Err }

func decodeListWorkflowsRequest(r *http.Request) (interface{}, error) {
	return nil, nil
}

// ListWorkflowsResponse is the response struct for a ListWorkflows request
type ListWorkflowsResponse struct {
	workflowList []Workflow
	Err          error `json:"error,omitempty"`
}

func (r ListWorkflowsResponse) error() error { return r.Err }

func (r ListWorkflowsResponse) encodeList(w http.ResponseWriter) error {
	jsn, err := json.MarshalIndent(r.workflowList, "", "  ")
	if err != nil {
		return err
	}
	w.Write(jsn)
	return nil
}

type listEncoder interface {
	encodeList(w http.ResponseWriter) error
}

// errorer is implemented by all concrete response types. It allows us to
// change the HTTP response code without needing to trigger an endpoint
// (transport-level) error. For more information, read the big comment in
// endpoint.go.
type errorer interface {
	error() error
}

// encodeResponse is the common method to encode all response types to the
// client. I chose to do it this way because I didn't know if something more
// specific was necessary. It's certainly possible to specialize on a
// per-response (per-method) basis.
func encodeResponse(w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(w, e.error())
		return nil
	}
	// for success responses
	if e, ok := response.(statuser); ok {
		w.WriteHeader(e.status())
	}

	// check if this is a collection
	if e, ok := response.(listEncoder); ok {
		return e.encodeList(w)

	}

	jsn, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return err
	}
	w.Write(jsn)
	return nil
}

func encodeError(w http.ResponseWriter, err error) {
	w.WriteHeader(codeFrom(err))
	response := map[string]interface{}{
		"error": err.Error(),
	}
	jsn, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Println(err)
		return
	}
	w.Write(jsn)
}

func codeFrom(err error) int {
	switch err {
	default:
		return http.StatusInternalServerError
	}
}

// ENDPOINTS
func makeNewWorkflowEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NewWorkflowRequest)
		if req.Workflow == nil {
			return NewWorkflowResponse{Err: ErrEmptyRequest}, nil
		}
		wfRequest := req.Workflow
		if wfRequest.Name == "" {
			return NewWorkflowResponse{Err: ErrEmptyRequest}, nil
		}
		workflow, err := svc.CreateWorkflow(wfRequest)
		if err != nil {
			return NewWorkflowResponse{Err: err}, nil
		}
		return NewWorkflowResponse{Workflow: workflow}, nil
	}
}

func makeListWorkflowsEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		workflows, err := svc.ListWorkflows()
		if err != nil {
			return ListWorkflowsResponse{Err: err}, nil
		}
		return ListWorkflowsResponse{workflowList: workflows}, nil
	}
}
