package workflow

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
)

// sql statements
var (
	createWorkflowStmt = `INSERT INTO workflows (name) VALUES ($1) 
						 ON CONFLICT ON CONSTRAINT workflows_name_key DO NOTHING
						 RETURNING workflow_uuid;`
	selectWorkflowsStmt        = `SELECT workflow_uuid, name FROM workflows`
	getProfilesForWorkflowStmt = `SELECT profiles.profile_uuid,payload_identifier FROM profiles 
								  LEFT JOIN workflow_profile 
								  ON workflow_profile.profile_uuid = profiles.profile_uuid 
								  WHERE workflow_profile.workflow_uuid=$1`
)

// WrkflowUUID is a filter we can add as a parameter to narrow down the list of returned results
type WrkflowUUID struct {
	UUID string
}

func (p WrkflowUUID) where() string {
	return fmt.Sprintf("workflow_uuid = '%s'", p.UUID)
}

// Create stores a new workflow in Postgres
func (store pgStore) CreateWorkflow(wf *Workflow) (*Workflow, error) {
	err := store.QueryRow(createWorkflowStmt, wf.Name).Scan(&wf.UUID)
	if err == sql.ErrNoRows {
		return nil, ErrExists
	}
	if err != nil {
		return nil, errors.Wrap(err, "pgStore create workflow")
	}
	profiles := wf.Profiles
	if err := store.addProfiles(wf.UUID, profiles...); err != nil {
		return nil, err
	}
	return wf, nil
}

func (store pgStore) UpdateWorkflow(wf *Workflow) (*Workflow, error) {
	if wf.UUID == "" || wf.Name == "" {
		return nil, errors.New("workflow must have UUID or name to be updated")
	}
	wfs, err := store.Workflows(WrkflowUUID{wf.UUID})
	if err != nil {
		return nil, err
	}
	if len(wfs) == 0 {
		return nil, errors.New("workflow not found in datastore")
	}

	retWf := wfs[0] //returned workflow
	if err := store.updateProfiles(wf, &retWf); err != nil {
		return nil, err
	}
	retWf.Profiles = wf.Profiles
	return &retWf, nil
}

func (store pgStore) updateProfiles(updated, inDatastore *Workflow) error {
	// if retWf has some profiles which are missing in the update request
	// remove them
	for _, p := range inDatastore.Profiles {
		if !updated.HasProfile(p.PayloadIdentifier) {
			if err := store.removeProfile(inDatastore.UUID, p.UUID); err != nil {
				return err
			}
		}
	}

	// if retWF is missing a profile, add it
	for _, p := range updated.Profiles {
		if !inDatastore.HasProfile(p.PayloadIdentifier) {
			if err := store.addProfile(inDatastore.UUID, p.UUID); err != nil {
				return errors.Wrap(err, "update wf, add profile")
			}
		}
	}
	return nil
}

func (store pgStore) addProfiles(wfUUID string, profiles ...Profile) error {
	if len(profiles) == 0 {
		return nil
	}
	for _, prf := range profiles {
		if err := store.addProfile(wfUUID, prf.UUID); err != nil {
			return err
		}
	}
	return nil
}

func (store pgStore) addProfile(wfUUID, pfUUID string) error {
	addProfileStmt := `INSERT INTO workflow_profile (workflow_uuid, profile_uuid) VALUES ($1, $2)
								  ON CONFLICT ON CONSTRAINT workflow_profile_pkey DO NOTHING;`

	_, err := store.Exec(addProfileStmt, wfUUID, pfUUID)
	if err != nil {
		return errors.Wrap(err, "pgStore add profile to workflow")
	}
	return nil
}

func (store pgStore) removeProfile(wfUUID, pfUUID string) error {
	rmProfileStmt := fmt.Sprintf(
		"DELETE FROM workflow_profile WHERE workflow_uuid = '%s' AND profile_uuid = '%s'",
		wfUUID,
		pfUUID,
	)
	_, err := store.Exec(rmProfileStmt)
	if err != nil {
		return errors.Wrap(err, "remove profile from workflow")
	}
	return nil
}

func (store pgStore) Workflows(params ...interface{}) ([]Workflow, error) {
	stmt := selectWorkflowsStmt
	stmt = addWhereFilters(stmt, params...)
	var workflows []Workflow
	err := store.Select(&workflows, stmt)
	if err != nil {
		return nil, errors.Wrap(err, "pgStore Workflows")
	}

	var withProfiles []Workflow
	for _, wf := range workflows {
		wf.Profiles, err = store.findProfilesForWorkflow(wf.UUID)
		if err != nil {
			return nil, err
		}
		withProfiles = append(withProfiles, wf)
	}
	return withProfiles, nil
}

func (store pgStore) findProfilesForWorkflow(wfUUID string) ([]Profile, error) {
	var profiles []Profile
	err := store.Select(&profiles, getProfilesForWorkflowStmt, wfUUID)
	if err != nil {
		return nil, err
	}
	return profiles, nil
}
