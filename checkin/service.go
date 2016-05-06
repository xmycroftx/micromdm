package checkin

import (
	"errors"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/micromdm/mdm"
	"github.com/micromdm/micromdm/checkin/push"
	"github.com/micromdm/micromdm/device"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

// ErrInvalidMessageType is an invalid checking command
var ErrInvalidMessageType = errors.New("invalid message type")

// MDMCheckinService models Apple's MDM Checkin commands
type MDMCheckinService interface {
	Authenticate(mdm.CheckinCommand) error
	TokenUpdate(mdm.CheckinCommand) error
	Checkout(mdm.CheckinCommand) error
	Enroll(udid string) (*device.Profile, error)
}

// NewCheckinService creates a new MDM Checkin Service
func NewCheckinService(options ...func(*config) error) MDMCheckinService {
	conf := &config{}
	defaultLogger := log.NewLogfmtLogger(os.Stderr)
	for _, option := range options {
		if err := option(conf); err != nil {
			defaultLogger.Log("err", err)
			os.Exit(1)
		}
	}
	var svc MDMCheckinService
	// p := mdmpush.New(conf.logger, conf.pushcert, conf.pushpass)
	// svc = mdmCheckinServie{pushsvc: p, db: conf.db}
	svc = mdmCheckinService{db: conf.db}
	if conf.logger != nil {
		svc = loggingMiddleware{conf.logger, svc}
	}

	fieldKeys := []string{"MessageType", "error"}
	requestCount := kitprometheus.NewCounter(stdprometheus.CounterOpts{
		Name: "request_count",
		Help: "http request count",
	}, fieldKeys)
	requestLatency := metrics.NewTimeHistogram(time.Microsecond, kitprometheus.NewSummary(stdprometheus.SummaryOpts{
		Name: "request_latency",
		Help: "http request duration",
	}, fieldKeys))
	svc = instrumentingMiddleware{requestCount, requestLatency, svc} // add metrics
	return svc
}

// Logger adds a logger to the service
func Logger(logger log.Logger) func(*config) error {
	return func(c *config) error {
		c.logger = logger
		return nil
	}
}

// Datastore adds a db connection to the service
func Datastore(db device.Datastore) func(*config) error {
	return func(c *config) error {
		c.db = db
		return nil
	}
}

//Push creates a push client
func Push(cert, password string) func(*config) error {
	return func(c *config) error {
		c.pushcert = cert
		c.pushpass = password
		return nil
	}
}

type config struct {
	pushcert string //path to cert
	pushpass string //password for cert
	logger   log.Logger
	db       device.Datastore
}

type mdmCheckinService struct {
	pushsvc mdmpush.Pusher
	db      device.Datastore
}

func (svc mdmCheckinService) Authenticate(cmd mdm.CheckinCommand) error {
	dev := &device.Device{
		UDID:         cmd.UDID,
		SerialNumber: &cmd.SerialNumber,
		OSVersion:    &cmd.OSVersion,
		BuildVersion: &cmd.BuildVersion,
		ProductName:  &cmd.ProductName,
		IMEI:         &cmd.IMEI,
		MEID:         &cmd.MEID,
		MDMTopic:     &cmd.Topic,
	}
	return svc.db.AddDevice(dev)
}

func (svc mdmCheckinService) TokenUpdate(cmd mdm.CheckinCommand) error {
	token := cmd.Token.String()
	unlockToken := cmd.UnlockToken.String()
	existing, err := svc.db.GetDeviceByUDID(cmd.UDID)
	if err != nil {
		return err
	}
	existing.Token = &token
	existing.MDMTopic = &cmd.Topic
	existing.PushMagic = &cmd.PushMagic
	existing.UnlockToken = &unlockToken
	existing.AwaitingConfiguration = &cmd.AwaitingConfiguration
	existing.Enrolled = boolPtr(true)
	err = svc.db.SaveDevice(existing)
	if err != nil {
		return err
	}
	// svc.pushsvc.Push(cmd.PushMagic, token)
	return nil
}

func (svc mdmCheckinService) Checkout(cmd mdm.CheckinCommand) error {
	existing, err := svc.db.GetDeviceByUDID(cmd.UDID)
	if err != nil {
		return err
	}
	existing.Enrolled = boolPtr(false)
	err = svc.db.SaveDevice(existing)
	if err != nil {
		return err
	}
	return nil
}

func (svc mdmCheckinService) Enroll(udid string) (*device.Profile, error) {
	profile, err := svc.db.GetProfileForDevice(udid)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

// return a pointer to a boolean
func boolPtr(b bool) *bool {
	return &b
}

// ServiceHandler creates an http handler
func ServiceHandler(ctx context.Context, svc MDMCheckinService) http.Handler {
	// endpoint
	checkin := makeCheckinEndpoint(svc)

	// handler
	checkinHandler := httptransport.NewServer(
		ctx,
		checkin,
		decodeMDMCheckinRequest,
		encodeResponse,
	)

	enroll := makeEnrollmentEndpoint(svc)

	enrollmentHandler := httptransport.NewServer(
		ctx,
		enroll,
		decodeMDMEnrollmentRequest,
		enrollResponse,
	)

	r := mux.NewRouter()
	r.Methods("PUT").Path("/mdm/checkin").Handler(checkinHandler)
	r.Methods("POST").Path("/mdm/checkin").Handler(enrollmentHandler)
	return r
}
