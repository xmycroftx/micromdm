package checkin

import (
	"errors"

	"github.com/go-kit/kit/endpoint"
	"github.com/micromdm/mdm"
	"golang.org/x/net/context"
)

// ErrInvalidMessageType is an invalid checking command
var errInvalidMessageType = errors.New("invalid message type")

type mdmCheckinRequest struct {
	mdm.CheckinCommand
}

type mdmCheckinResponse struct {
	Err error `plist:"error,omitempty"`
}

func (r mdmCheckinResponse) error() error { return r.Err }

func makeCheckinEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(mdmCheckinRequest)
		var err error
		switch req.MessageType {
		case "Authenticate":
			err = svc.Authenticate(req.CheckinCommand)
		case "TokenUpdate":
			err = svc.TokenUpdate(req.CheckinCommand)
		case "CheckOut":
			err = svc.Checkout(req.CheckinCommand)
		default:
			return mdmCheckinResponse{errInvalidMessageType}, nil
		}
		if err != nil {
			return mdmCheckinResponse{err}, nil
		}
		return mdmCheckinResponse{}, nil
	}
}

type depEnrollmentRequest struct {
	mdm.DEPEnrollmentRequest
}

type depEnrollmentResponse struct {
	Profile []byte // MDM Enrollment Profile
	Err     error  `plist:"error,omitempty"`
}

func (r depEnrollmentResponse) error() error { return r.Err }
