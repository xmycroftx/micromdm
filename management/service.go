package management

import (
	"github.com/micromdm/dep"
	"github.com/micromdm/micromdm/device"
	"github.com/micromdm/micromdm/workflow"
	"github.com/pkg/errors"
)

// ErrNotFound ...
var ErrNotFound = errors.New("not found")

// Service is the interface that provides methods for managing devices
type Service interface {
	AddProfile(prf *workflow.Profile) (*workflow.Profile, error)
	Profiles() ([]workflow.Profile, error)
	Profile(uuid string) (*workflow.Profile, error)
	DeleteProfile(uuid string) error
	FetchDEPDevices() error
}

type service struct {
	depClient dep.Client
	devices   device.Datastore
	workflows workflow.Datastore
}

func (svc service) AddProfile(prf *workflow.Profile) (*workflow.Profile, error) {
	return svc.workflows.CreateProfile(prf)
}

func (svc service) Profiles() ([]workflow.Profile, error) {
	return svc.workflows.Profiles()
}

// Profile returns a single profile given an UUID
func (svc service) Profile(uuid string) (*workflow.Profile, error) {
	profiles, err := svc.workflows.Profiles(workflow.ProfileUUID{uuid})
	if err != nil {
		return nil, err
	}
	if len(profiles) == 0 {
		return nil, ErrNotFound
	}
	pf := profiles[0]
	return &pf, nil
}

func (svc service) DeleteProfile(uuid string) error {
	pr, err := svc.Profile(uuid) // get profile from datastore
	if err != nil {
		return err
	}
	err = svc.workflows.DeleteProfile(pr)
	if err != nil {
		return err
	}
	return nil
}

func (svc service) FetchDEPDevices() error {
	fetched, err := svc.depClient.FetchDevices(dep.Limit(100))
	if err != nil {
		return errors.Wrap(err, "management: dep fetch")
	}
	for _, d := range fetched.Devices {
		dev := device.NewFromDEP(d)
		_, err := svc.devices.New("fetch", dev)
		if err != nil {
			return errors.Wrap(err, "management: dep fetch")
		}
	}
	return nil
}

// NewService creates a management service
func NewService(ds device.Datastore, ws workflow.Datastore, dc dep.Client) Service {
	return &service{
		devices:   ds,
		depClient: dc,
		workflows: ws,
	}
}
