package checkin

import "github.com/micromdm/mdm"

type mdmCheckinRequest struct {
	mdm.CheckinCommand
}

type mdmCheckinResponse struct {
	Err error `plist:"error,omitempty"`
}

func (r mdmCheckinResponse) error() error { return r.Err }

type depEnrollmentRequest struct {
	mdm.DEPEnrollmentRequest
}

type depEnrollmentResponse struct {
	Profile []byte // MDM Enrollment Profile
	Err     error  `plist:"error,omitempty"`
}

func (r depEnrollmentResponse) error() error { return r.Err }
