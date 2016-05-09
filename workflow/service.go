package workflow

import (
	"net/http"
	"os"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

// Service manages workflows
type Service interface {
	CreateWorkflow(*Workflow) (*Workflow, error)
	ListWorkflows() ([]Workflow, error)
}

type workflowService struct {
	info  log.Logger
	debug log.Logger
	db    Datastore
}

func (svc workflowService) CreateWorkflow(wfRequest *Workflow) (*Workflow, error) {
	svc.debug.Log("action", "CreateWorkflow", "name", wfRequest.Name)
	wf, err := svc.db.CreateWorkflow(wfRequest.Name)
	if err != nil {
		return nil, errors.Wrap(err, "workflow service")
	}
	// the request had no profiles, return response
	if len(wfRequest.Profiles) == 0 {
		return wf, nil
	}
	// if this loop fails the workflow already exists
	// need to delete the workflow? ugh
	// also, the PayloadIdentifier is not validated
	// it should be
	for _, pf := range wfRequest.Profiles {
		if err := svc.db.AddProfile(wf.UUID, pf.UUID); err != nil {
			return nil, errors.Wrap(err, "createworkflow failed to add profile")
		}
		wf.Profiles = append(wf.Profiles, configProfile{
			UUID:              pf.UUID,
			PayloadIdentifier: pf.PayloadIdentifier,
		})
	}
	return wf, nil
}

func (svc workflowService) ListWorkflows() ([]Workflow, error) {
	svc.debug.Log("Listing Workflows")
	return svc.db.GetWorkflows()
}

// NewService creates a new MDM Command Service
func NewService(options ...func(*config) error) Service {
	conf := &config{}
	defaultLogger := log.NewLogfmtLogger(os.Stderr)
	for _, option := range options {
		if err := option(conf); err != nil {
			defaultLogger.Log("err", err)
			os.Exit(1)
		}
	}
	var svc Service
	svc = workflowService{
		info:  infoLogger(conf),
		debug: debugLogger(conf),
		db:    conf.db,
	}
	return svc
}

// DB adds a db connection to the service
func DB(db Datastore) func(*config) error {
	return func(c *config) error {
		c.db = db
		return nil
	}
}

// ServiceHandler returns an http handler for the command service
func ServiceHandler(ctx context.Context, svc Service) http.Handler {
	commonOptions := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(encodeError),
	}
	newWorkflowEndpoint := makeNewWorkflowEndpoint(svc)
	newWorkflowHandler := httptransport.NewServer(
		ctx,
		newWorkflowEndpoint,
		decodeNewWorkflowRequest,
		encodeResponse,
		commonOptions...,
	)

	listWorkflowsHandler := httptransport.NewServer(
		ctx,
		makeListWorkflowsEndpoint(svc),
		decodeListWorkflowsRequest,
		encodeResponse,
		commonOptions...,
	)
	r := mux.NewRouter()

	r.Handle("/mdm/workflows", newWorkflowHandler).Methods("POST")
	r.Handle("/mdm/workflows", listWorkflowsHandler).Methods("GET")
	return r
}
