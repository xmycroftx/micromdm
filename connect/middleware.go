package connect

import (
	"github.com/micromdm/micromdm/device"
	"golang.org/x/net/context"
)

type lastCheckinMiddleware struct {
	devices device.Datastore
	next    Service
}

func NewMiddleware(datastore device.Datastore, next Service) (*lastCheckinMiddleware) {
	return &lastCheckinMiddleware{
		datastore,
		next,
	}
}

func (mw lastCheckinMiddleware) Acknowledge(ctx context.Context, req mdmConnectRequest) (int, error) {
	mw.devices.UpdateDeviceCheckinByUDID(req.UDID)
	return mw.next.Acknowledge(ctx, req)
}

func (mw lastCheckinMiddleware) NextCommand(ctx context.Context, req mdmConnectRequest) ([]byte, int, error) {
	mw.devices.UpdateDeviceCheckinByUDID(req.UDID)
	return mw.next.NextCommand(ctx, req)
}
