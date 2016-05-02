package profile

import "time"

// Profile is a configuration profile
// See https://developer.apple.com/library/ios/featuredarticles/iPhoneConfigurationProfileRef/Introduction/Introduction.html#//apple_ref/doc/uid/TP40010206-CH1-SW4
type Profile struct {
	PayloadContent           []PayloadDictionary
	PayloadDescription       string    `plist:",omitempty" json:",omitempty"`
	PayloadDisplayName       string    `plist:",omitempty" json:",omitempty"`
	PayloadExpirationDate    time.Time `plist:",omitempty" json:",omitempty"`
	PayloadIdentifier        string
	PayloadOrganization      string `plist:",omitempty" json:",omitempty"`
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
