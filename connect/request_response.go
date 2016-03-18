package connect

import "github.com/micromdm/mdm"

type mdmConnectRequest struct {
	mdm.Response
}

type mdmConnectResponse struct {
	payload []byte
	Err     error `plist:"error,omitempty"`
}

func (r mdmConnectResponse) error() error { return r.Err }
