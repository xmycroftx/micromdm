package checkin

import (
	"github.com/go-kit/kit/endpoint"
	"golang.org/x/net/context"
)

func makeCheckinEndpoint(svc MDMCheckinService) endpoint.Endpoint {
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
			return mdmCheckinResponse{ErrInvalidMessageType}, nil
		}
		if err != nil {
			return mdmCheckinResponse{err}, nil
		}
		return mdmCheckinResponse{}, nil
	}
}

func makeEnrollmentEndpoint(svc MDMCheckinService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(depEnrollmentRequest)
		profile, err := svc.Enroll(req.DEPEnrollmentRequest.UDID)
		if err != nil {
			return depEnrollmentResponse{Err: err}, nil
		}
		return depEnrollmentResponse{Profile: []byte(*profile)}, nil
	}
}
