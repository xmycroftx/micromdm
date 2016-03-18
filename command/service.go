package command

import (
	"net/http"
	"os"

	"golang.org/x/net/context"

	httptransport "github.com/go-kit/kit/transport/http"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/micromdm/mdm"
)

// MDMCommandService allows creating and deleting MDM Command Payloads
type MDMCommandService interface {
	NewCommand(*mdm.CommandRequest) (*mdm.Payload, error)
	NextCommand(udid string) ([]byte, int, error)
	DeleteCommand(deviceUDID, commandUUID string) (int, error)
}

type mdmCommandService struct {
	// a redis datastore
	db Datastore
}

type config struct {
	logger log.Logger
	db     Datastore
}

// NewCommandService creates a new MDM Command Service
func NewCommandService(options ...func(*config) error) MDMCommandService {
	conf := &config{}
	defaultLogger := log.NewLogfmtLogger(os.Stderr)
	for _, option := range options {
		if err := option(conf); err != nil {
			defaultLogger.Log("err", err)
			os.Exit(1)
		}
	}
	var svc MDMCommandService
	svc = mdmCommandService{db: conf.db}
	return svc
}

// Logger adds a logger to the service
func Logger(logger log.Logger) func(*config) error {
	return func(c *config) error {
		c.logger = logger
		return nil
	}
}

// DB adds a db connection to the service
func DB(db Datastore) func(*config) error {
	return func(c *config) error {
		c.db = db
		return nil
	}
}

// ServiceHandler returns an http handler for the command service
func ServiceHandler(ctx context.Context, svc MDMCommandService) http.Handler {
	commonOptions := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(encodeError),
	}
	newCommandEndpoint := makeNewCommandEndpoint(svc)
	newCommandHandler := httptransport.NewServer(
		ctx,
		newCommandEndpoint,
		decodeNewCommandRequest,
		encodeResponse,
		commonOptions...,
	)
	nextCommandEndpoint := makeNextCommandEndpoint(svc)
	nextCommandHandler := httptransport.NewServer(
		ctx,
		nextCommandEndpoint,
		decodeNextCommandRequest,
		encodeResponse,
		commonOptions...,
	)
	deleteCommandEndpoint := makeDeleteCommandEndpoint(svc)
	deleteCommandHandler := httptransport.NewServer(
		ctx,
		deleteCommandEndpoint,
		decodeDeleteCommandRequest,
		encodeResponse,
		commonOptions...,
	)
	r := mux.NewRouter()
	r.Methods("POST").Path("/mdm/commands").Handler(newCommandHandler)
	r.Methods("GET").Path("/mdm/commands/{udid}/next").Handler(nextCommandHandler)
	r.Methods("DELETE").Path("/mdm/commands/{udid}/{uuid}").Handler(deleteCommandHandler)
	return r
}

func (svc mdmCommandService) NewCommand(request *mdm.CommandRequest) (*mdm.Payload, error) {
	// create a payload
	payload, err := mdm.NewPayload(request)
	if err != nil {
		return nil, err
	}
	// save in redis
	err = svc.db.SavePayload(payload)
	if err != nil {
		return nil, err
	}
	// add command to a queue in redis
	err = svc.db.QueueCommand(request.UDID, payload.CommandUUID)
	if err != nil {
		return nil, err
	}
	// return created payload to user
	return payload, nil
}

// NextCommand returns an MDM Payload from a list of queued payloads
func (svc mdmCommandService) NextCommand(udid string) ([]byte, int, error) {
	return svc.db.NextCommand(udid)
}

// DeleteCommand returns an MDM Payload from a list of queued payloads
func (svc mdmCommandService) DeleteCommand(deviceUDID, commandUUID string) (int, error) {
	return svc.db.DeleteCommand(deviceUDID, commandUUID)
}
