package profile

import (
	"net/http"
	"os"

	httptransport "github.com/go-kit/kit/transport/http"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

// Service defines the things you can do with a Profile
type Service interface {
	AddProfile(*Profile) (*Profile, error)
	ListProfiles() ([]Profile, error)
}

type profileService struct {
	info  log.Logger
	debug log.Logger
	db    Datastore
}

func (svc profileService) AddProfile(pf *Profile) (*Profile, error) {
	svc.debug.Log("action", "AddProfile", "identifier", pf.PayloadIdentifier)
	return svc.db.AddProfile(pf)
}

func (svc profileService) ListProfiles() ([]Profile, error) {
	svc.debug.Log("action", "ListProfiles")
	return svc.db.GetProfiles()
}

// NewService creates a new Configuration Profile Service
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
	svc = profileService{
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
	addProfileEndpoint := makeAddProfileEndpoint(svc)
	addProfileHandler := httptransport.NewServer(
		ctx,
		addProfileEndpoint,
		decodeAddProfileRequest,
		encodeResponse,
		commonOptions...,
	)

	listProfilesEndpoint := makeListProfilesEndpoint(svc)
	listProfilesHandler := httptransport.NewServer(
		ctx,
		listProfilesEndpoint,
		decodeListProfileRequest,
		encodeResponse,
		commonOptions...,
	)
	r := mux.NewRouter()
	r.Methods("POST").Path("/mdm/profiles").Handler(addProfileHandler)
	r.Methods("GET").Path("/mdm/profiles").Handler(listProfilesHandler)
	return r
}
