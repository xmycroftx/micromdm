package workflow

import "errors"

// ErrExists is returned if a workflow already exists
var ErrExists = errors.New("workflow already exists. each workflow must have a unique name")

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
