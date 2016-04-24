package connect

import (
	"errors"

	"github.com/go-kit/kit/endpoint"
	"golang.org/x/net/context"
)

// ErrInvalidMessageType is an invalid checking command
var ErrInvalidMessageType = errors.New("Invalid MessageType")

func makeConnectEndpoint(svc MDMConnectService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(mdmConnectRequest)
		if req.UserID != nil {
			// don't handle user
			return mdmConnectResponse{}, nil
		}
		var err error
		switch req.Status {
		case "Acknowledged":
			total, err := svc.Acknowledge(req.UDID, req.CommandUUID)
			if err != nil {
				return mdmConnectResponse{Err: err}, nil
			}
			if total != 0 {
				next, _, err := svc.NextCommand(req.UDID)
				if err != nil {
					return mdmConnectResponse{Err: err}, nil
				}
				return mdmConnectResponse{payload: next}, nil
			}
		case "Idle":
			next, total, err := svc.NextCommand(req.UDID)
			if err != nil {
				return mdmConnectResponse{Err: err}, nil
			}
			if total == 0 {
				return mdmConnectResponse{}, nil
			}
			return mdmConnectResponse{payload: next}, nil
		default:
			return mdmConnectResponse{Err: ErrInvalidMessageType}, nil
		}
		if err != nil {
			return mdmConnectResponse{Err: err}, nil
		}
		return mdmConnectResponse{}, nil
	}
}
