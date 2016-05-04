package profile

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"golang.org/x/net/context"
)

var (
	// ErrEmptyRequest is returned if the request body is empty
	ErrEmptyRequest = errors.New("request must contain a profile identifier")
	errBadRouting   = errors.New("inconsistent mapping between route and handler (programmer error)")
)

// -- AddProfile

// AddProfileRequest is a request to add a configuration profile
type AddProfileRequest struct {
	*Profile
}

func decodeAddProfileRequest(r *http.Request) (interface{}, error) {
	var request AddProfileRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	return request, err
}

// AddProfileResponse is a response to a AddProfile request
type AddProfileResponse struct {
	*Profile
	Err error `json:"error,omitempty"`
}

func (r AddProfileResponse) status() int  { return 201 }
func (r AddProfileResponse) error() error { return r.Err }

// -- ListProfiles

// // ListProfilesRequest ...
// type ListProfilesRequest struct{}

func decodeListProfileRequest(r *http.Request) (interface{}, error) {
	return nil, nil
}

// type profileList []Profile

// ListProfilesResponse ...
type ListProfilesResponse struct {
	profileList []Profile
	Err         error `json:"error,omitempty"`
}

func (r ListProfilesResponse) error() error { return r.Err }

func (r ListProfilesResponse) encodeList(w http.ResponseWriter) error {
	jsn, err := json.MarshalIndent(r.profileList, "", "  ")
	if err != nil {
		return err
	}
	w.Write(jsn)
	return nil
}

// ENDPOINTS
func makeAddProfileEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AddProfileRequest)
		if req.Profile == nil {
			return AddProfileResponse{Err: ErrEmptyRequest}, nil
		}
		if req.PayloadIdentifier == "" {
			return AddProfileResponse{Err: ErrEmptyRequest}, nil
		}
		pf, err := svc.AddProfile(req.Profile)
		if err != nil {
			return AddProfileResponse{Err: err}, nil
		}
		return AddProfileResponse{Profile: pf}, nil
	}
}

func makeListProfilesEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		profiles, err := svc.ListProfiles()
		if err != nil {
			return ListProfilesResponse{Err: err}, nil
		}
		// list := profileList(profiles)
		return ListProfilesResponse{profileList: profiles}, nil
	}
}

// for API statuses other than 200OK
type statuser interface {
	status() int
}

type errorer interface {
	error() error
}

type listEncoder interface {
	encodeList(w http.ResponseWriter) error
}

// encodeResponse is the common method to encode all response types to the
// client. I chose to do it this way because I didn't know if something more
// specific was necessary. It's certainly possible to specialize on a
// per-response (per-method) basis.
func encodeResponse(w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(w, e.error())
		return nil
	}
	// for success responses
	if e, ok := response.(statuser); ok {
		w.WriteHeader(e.status())
	}

	// check if this is a collection
	if e, ok := response.(listEncoder); ok {
		return e.encodeList(w)

	}

	jsn, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return err
	}
	w.Write(jsn)
	return nil
}

func encodeError(w http.ResponseWriter, err error) {
	w.WriteHeader(codeFrom(err))
	response := map[string]interface{}{
		"error": err.Error(),
	}
	jsn, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Println(err)
		return
	}
	w.Write(jsn)
}

func codeFrom(err error) int {
	switch err {
	case ErrExists:
		return http.StatusConflict
	case ErrEmptyRequest:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
