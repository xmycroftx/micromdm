package device

// AssignWorkflowRequest is a request for AssignWorkflow
type assignWorkflowRequest struct {
	WorkflowUUID string `json:"workflow_uuid"`
	DeviceUDID   string `json:"-"`
}

type assignWorkflowResponse struct {
	Err error `json:"error,omitempty"`
}

func (r assignWorkflowResponse) error() error { return r.Err }
