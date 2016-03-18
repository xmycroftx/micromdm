package connect

import (
	"net/http"
	"os"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/micromdm/micromdm/command"
)

// MDMConnectService ...
type MDMConnectService interface {
	Acknowledge(deviceUDID, commandUUID string) (int, error)
	NextCommand(deviceUDID string) ([]byte, int, error)
}

type mdmConnectService struct {
	redis command.Datastore
}

func (svc mdmConnectService) Acknowledge(deviceUDID, commandUUID string) (int, error) {
	return svc.redis.DeleteCommand(deviceUDID, commandUUID)

}

func (svc mdmConnectService) NextCommand(deviceUDID string) ([]byte, int, error) {
	return svc.redis.NextCommand(deviceUDID)
}

type config struct {
	logger log.Logger
	redis  command.Datastore
}

// NewConnectService creates a new MDM Connect Service
func NewConnectService(options ...func(*config) error) MDMConnectService {
	conf := &config{}
	defaultLogger := log.NewLogfmtLogger(os.Stderr)
	for _, option := range options {
		if err := option(conf); err != nil {
			defaultLogger.Log("err", err)
			os.Exit(1)
		}
	}
	var svc MDMConnectService
	svc = mdmConnectService{conf.redis}
	if conf.logger != nil {
		// svc = loggingMiddleware{conf.logger, svc}
	}

	return svc
}

// Redis adds a db connection to the service
func Redis(db command.Datastore) func(*config) error {
	return func(c *config) error {
		c.redis = db
		return nil
	}
}

// ServiceHandler creates an http handler
func ServiceHandler(ctx context.Context, svc MDMConnectService) http.Handler {
	// endpoint
	connect := makeConnectEndpoint(svc)

	// handler
	connectHandler := httptransport.NewServer(
		ctx,
		connect,
		decodeMDMConnectRequest,
		encodeResponse,
	)
	return connectHandler
}
