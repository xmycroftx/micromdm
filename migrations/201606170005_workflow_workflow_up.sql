CREATE TABLE IF NOT EXISTS workflow_workflow (
  workflow_uuid uuid REFERENCES workflows,
  included_workflow_uuid uuid REFERENCES workflows(workflow_uuid),
  PRIMARY KEY (workflow_uuid, included_workflow_uuid)
);

