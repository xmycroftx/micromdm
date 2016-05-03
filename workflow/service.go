package workflow

import (
	"net/http"
	"os"

	httptransport "github.com/go-kit/kit/transport/http"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

// Service manages workflows
type Service interface {
	CreateWorkflow(string) (*Workflow, error)
	ListWorkflows() ([]Workflow, error)
}

type workflowService struct {
	info  log.Logger
	debug log.Logger
	db    Datastore
}

func (svc workflowService) CreateWorkflow(name string) (*Workflow, error) {
	svc.debug.Log("action", "CreateWorkflow", "name", name)
	return svc.db.CreateWorkflow(name)
}

func (svc workflowService) ListWorkflows() ([]Workflow, error) {
	svc.debug.Log("Listing Workflows")
	return nil, nil
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
	r := mux.NewRouter()
	r.Methods("POST").Path("/mdm/workflows").Handler(newWorkflowHandler)
	return r
}
