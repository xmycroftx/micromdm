// Package profile manages configuration profile payloads
package profile

import (
	"time"

	"github.com/go-kit/kit/log"
)

// Profile is a configuration profile
type Profile struct {
	UUID              string `plist:"-" json:"profile_uuid,omitempty" db:"profile_uuid"`
	PayloadIdentifier string `json:"payload_identifier" db:"identifier"`
	Data              string `json:"data,omitempty" db:"data"`
}

// Logger adds a logger to the database config
func Logger(logger log.Logger) func(*config) error {
	return func(c *config) error {
		c.logger = logger
		return nil
	}
}

// Debug adds a debug logger to the database config
func Debug() func(*config) error {
	return func(c *config) error {
		c.debug = true
		return nil
	}
}

// XMLProfile is a configuration profile
// See https://developer.apple.com/library/ios/featuredarticles/iPhoneConfigurationProfileRef/Introduction/Introduction.html#//apple_ref/doc/uid/TP40010206-CH1-SW4
// Ignore this for now
type XMLProfile struct {
	UUID                     string `plist:"-" json:"-" db:"profile_uuid"`
	PayloadContent           []PayloadDictionary
	PayloadDescription       string    `plist:",omitempty" json:",omitempty"`
	PayloadDisplayName       string    `plist:",omitempty" json:",omitempty"`
	PayloadExpirationDate    time.Time `plist:",omitempty" json:",omitempty"`
	PayloadIdentifier        string    `db:"identifier"`
	PayloadOrganization      string    `plist:",omitempty" json:",omitempty"`
	PayloadUUID              string
	PayloadRemovalDisallowed bool `plist:",omitempty" json:",omitempty"`
	PayloadType              string
	PayloadVersion           int
	PayloadScope             string    `plist:",omitempty" json:",omitempty"`
	RemovalDate              time.Time `plist:",omitempty" json:",omitempty"`
	DurationUntilRemoval     float64   `plist:",omitempty" json:",omitempty"`
	//ConsentText dict TODO
}

// PayloadDictionary is a configuration profile payload
type PayloadDictionary struct {
	PayloadType         string
	PayloadVersion      int
	PayloadIdentifier   string
	PayloadUUID         string
	PayloadDisplayName  string
	PayloadDescription  string
	PayloadOrganization string
}
