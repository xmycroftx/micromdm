CREATE TABLE IF NOT EXISTS workflow_profile (
  workflow_uuid uuid REFERENCES workflows,
  profile_uuid uuid REFERENCES profiles,
  PRIMARY KEY (workflow_uuid, profile_uuid)
);

