package management

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/micromdm/micromdm/workflow"
	"golang.org/x/net/context"
)

var (
	// ErrEmptyRequest is returned if the request body is empty
	errEmptyRequest = errors.New("request must contain a profile identifier")
	errBadRouting   = errors.New("inconsistent mapping between route and handler (programmer error)")
)

type fetchDEPDevicesRequest struct{}

type fetchDEPDevicesResponse struct {
	Err error `json:"error,omitempty"`
}

func (r fetchDEPDevicesResponse) error() error { return r.Err }

func makeFetchDevicesEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		err := svc.FetchDEPDevices()
		return fetchDEPDevicesResponse{Err: err}, nil
	}
}

type addProfileRequest struct {
	*workflow.Profile
}

type addProfileResponse struct {
	*workflow.Profile
	Err error `json:"error,omitempty"`
}

func (r addProfileResponse) status() int { return http.StatusCreated }

func (r addProfileResponse) error() error { return r.Err }

func makeAddProfileEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(addProfileRequest)
		pf, err := svc.AddProfile(req.Profile)
		return addProfileResponse{Err: err, Profile: pf}, nil
	}
}

type listProfilesRequest struct{}

type listProfilesResponse struct {
	profiles []workflow.Profile
	Err      error `json:"error,omitempty"`
}

func (r listProfilesResponse) error() error { return r.Err }

func (r listProfilesResponse) encodeList(w http.ResponseWriter) error {
	jsn, err := json.MarshalIndent(r.profiles, "", "  ")
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(jsn)
	return nil
}

func makeListProfilesEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		pf, err := svc.Profiles()
		return listProfilesResponse{Err: err, profiles: pf}, nil
	}
}

type showProfileRequest struct {
	UUID string
}

type showProfileResponse struct {
	*workflow.Profile
	Err error `json:"error,omitempty"`
}

func (r showProfileResponse) error() error { return r.Err }

func makeShowProfileEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(showProfileRequest)
		pf, err := svc.Profile(req.UUID)
		if err != nil {
			return showProfileResponse{Err: err}, nil
		}
		return showProfileResponse{Profile: pf}, nil
	}
}

type deleteProfileRequest struct {
	UUID string
}

type deleteProfileResponse struct {
	Err error `json:"error,omitempty"`
}

func (r deleteProfileResponse) status() int  { return http.StatusNoContent }
func (r deleteProfileResponse) error() error { return r.Err }

func makeDeleteProfileEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(deleteProfileRequest)
		err := svc.DeleteProfile(req.UUID)
		if err != nil {
			return deleteProfileResponse{Err: err}, nil
		}
		return deleteProfileResponse{}, nil
	}
}
