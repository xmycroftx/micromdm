package connect

import (
	"github.com/micromdm/mdm"
	"github.com/micromdm/micromdm/command"
	"github.com/micromdm/micromdm/device"
	"github.com/pkg/errors"
)

// Service defines methods for an MDM service
type Service interface {
	Acknowledge(deviceUDID, commandUUID string) (int, error)
	NextCommand(deviceUDID string) ([]byte, int, error)
}

// NewService creates a mdm service
func NewService(devices device.Datastore, cs command.Service) Service {
	return &service{
		commands: cs,
		devices:  devices,
	}
}

type service struct {
	devices  device.Datastore
	commands command.Service
}

func (svc service) Acknowledge(deviceUDID, commandUUID string) (int, error) {
	total, err := svc.commands.DeleteCommand(deviceUDID, commandUUID)
	if err != nil {
		return total, err
	}
	if total == 0 {
		total, err = svc.checkRequeue(deviceUDID)
		if err != nil {
			return total, err
		}
		return total, nil
	}
	return total, nil
}

func (svc service) NextCommand(deviceUDID string) ([]byte, int, error) {
	return svc.commands.NextCommand(deviceUDID)
}

func (svc service) checkRequeue(deviceUDID string) (int, error) {
	existing, err := svc.devices.GetDeviceByUDID(deviceUDID, []string{"awaiting_configuration"}...)
	if err != nil {
		return 0, errors.Wrap(err, "check and requeue")
	}
	if existing.AwaitingConfiguration {
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
