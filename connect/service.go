package connect

import (
	"fmt"
	"net/http"
	"os"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/micromdm/mdm"
	"github.com/micromdm/micromdm/command"
	"github.com/micromdm/micromdm/device"
	"github.com/pkg/errors"
)

// MDMConnectService ...
type MDMConnectService interface {
	Acknowledge(deviceUDID, commandUUID string) (int, error)
	NextCommand(deviceUDID string) ([]byte, int, error)
}

type mdmConnectService struct {
	redis    command.Datastore
	devices  device.Datastore
	commands command.MDMCommandService
}

func (svc mdmConnectService) Acknowledge(deviceUDID, commandUUID string) (int, error) {
	total, err := svc.redis.DeleteCommand(deviceUDID, commandUUID)
	if err != nil {
		return total, err
	}
	if total == 0 {
		fmt.Println("requeueing command")
		total, err = svc.checkRequeue(deviceUDID)
		if err != nil {
			return total, err
		}
		return total, nil
	}
	return total, nil
}

func (svc mdmConnectService) NextCommand(deviceUDID string) ([]byte, int, error) {
	return svc.redis.NextCommand(deviceUDID)
}

func (svc mdmConnectService) checkRequeue(deviceUDID string) (int, error) {
	existing, err := svc.devices.GetDeviceByUDID(deviceUDID)
	if err != nil {
		return 0, errors.Wrap(err, "check and requeue")
	}
	if *existing.AwaitingConfiguration {
		cmdRequest := &mdm.CommandRequest{
			UDID:        deviceUDID,
			RequestType: "DeviceConfigured",
		}

		_, err := svc.commands.NewCommand(cmdRequest)
		if err != nil {
			return 0, err
		}
		return 1, nil
	}
	return 0, nil
}

type config struct {
	logger   log.Logger
	redis    command.Datastore
	devices  device.Datastore
	commands command.MDMCommandService
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
	svc = mdmConnectService{redis: conf.redis, devices: conf.devices, commands: conf.commands}
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

// Devices adds a db connection to the service
func Devices(db device.Datastore) func(*config) error {
	return func(c *config) error {
		c.devices = db
		return nil
	}
}

// Commands adds a db connection to the service
func Commands(svc command.MDMCommandService) func(*config) error {
	return func(c *config) error {
		c.commands = svc
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
