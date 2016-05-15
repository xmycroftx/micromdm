package workflow

import "errors"

// ErrExists is returned when trying to add a resource which already exists
var ErrExists = errors.New("resource already exists in the datastore")

// Workflow describes a workflow that a device will execute
// A workflow contains a list of configuration profiles,
// Applications and included workflows
type Workflow struct {
	UUID     string    `json:"uuid" db:"workflow_uuid"`
	Name     string    `json:"name" db:"name"`
	Profiles []Profile `json:"profiles"`
	// Applications      []application
	// IncludedWorkflows []Workflow
}
