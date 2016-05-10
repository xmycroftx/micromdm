package device

import (
	"database/sql/driver"
	"errors"
	"time"
)

// Device represents an iOS or OS X Computer
type Device struct {
	// Primary key is UUID
	UUID         string `json:"uuid" db:"device_uuid"`
	UDID         string `json:"udid"`
	SerialNumber string `json:"serial_number,omitempty" db:"serial_number,omitempty"`
	OSVersion    string `json:"os_version,omitempty" db:"os_version,omitempty"`
	BuildVersion string `json:"build_version,omitempty" db:"build_version,omitempty"`
	ProductName  string `json:"product_name,omitempty" db:"product_name,omitempty"`
	IMEI         string `json:"imei,omitempty" db:"imei,omitempty"`
	MEID         string `json:"meid,omitempty" db:"meid,omitempty"`
	//Apple MDM Protocol Topic
	MDMTopic               string           `json:"mdm_topic,omitempty" db:"apple_mdm_topic,omitempty"`
	PushMagic              string           `json:"push_magic,omitempty" db:"apple_push_magic,omitempty"`
	AwaitingConfiguration  bool             `json:"awaiting_configuration,omitempty" db:"awaiting_configuration,omitempty"`
	Token                  string           `json:"token,omitempty" db:"apple_mdm_token,omitempty"`
	UnlockToken            string           `json:"unlock_token,omitempty" db:"unlock_token,omitempty"`
	Enrolled               bool             `json:"enrolled,omitempty" db:"mdm_enrolled,omitempty"`
	Workflow               string           `json:"workflow,omitempty" db:"workflow_uuid"`
	DEPDevice              bool             `json:"dep_device,omitempty" db:"dep_device,omitempty"`
	Description            string           `json:"description,omitempty" db:"description"`
	Model                  string           `json:"model,omitempty" db:"model"`
	Color                  string           `json:"color,omitempty" db:"color"`
	AssetTag               string           `json:"asset_tag,omitempty" db:"asset_tag"`
	DEPProfileStatus       DEPProfileStatus `json:"dep_profile_status,omitempty" db:"dep_profile_status"`
	DEPProfileUUID         string           `json:"dep_profile_uuid,omitempty" db:"dep_profile_uuid"`
	DEPProfileAssignTime   time.Time        `json:"dep_profile_assign_time,omitempty" db:"dep_profile_assign_time"`
	DEPProfilePushTime     time.Time        `json:"dep_profile_push_time,omitempty" db:"dep_profile_push_time"`
	DEPProfileAssignedDate time.Time        `json:"dep_profile_assigned_date" db:"dep_profile_assigned_date"`
	DEPProfileAssignedBy   string           `json:"dep_profile_assigned_by" db:"dep_profile_assigned_by"`
}

// DEPProfileStatus is the status of the DEP Profile
// can be either "empty", "assigned", "pushed", or "removed"
type DEPProfileStatus string

// DEPProfileStatus values
const (
	EMPTY    DEPProfileStatus = "empty"
	ASSIGNED                  = "assigned"
	PUSHED                    = "pushed"
	REMOVED                   = "removed"
)

// Value implements Valuer from database/sql
func (status DEPProfileStatus) Value() (driver.Value, error) { return string(status), nil }

// Scan implements Scanner from database/sql
func (status *DEPProfileStatus) Scan(value interface{}) error {
	if value == nil {
		*status = EMPTY
		return nil
	}
	if sv, err := driver.String.ConvertValue(value); err == nil {
		if v, ok := sv.(string); ok {
			*status = DEPProfileStatus(v)
			return nil
		}
	}
	// otherwise, return an error
	return errors.New("failed to scan DEPProfileStatus")

}
