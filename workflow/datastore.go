package workflow

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	kitlog "github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgres driver
	"github.com/pkg/errors"
)

// ErrExists is returned if a workflow already exists
var ErrExists = errors.New("workflow already exists. each workflow must have a unique name")

// Profile is configuration profile in a workflow
type Profile struct {
	UUID string
}

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

// Datastore manages interactions of workflows in a database
type Datastore interface {
	// Create adds a new workflow to the datastore
	Create(wf *Workflow) (*Workflow, error)
	Workflows(params ...interface{}) ([]Workflow, error)
}

type pgStore struct {
	*sqlx.DB
}

// Create stores a new workflow in Postgres
func (store pgStore) Create(wf *Workflow) (*Workflow, error) {
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

func (store pgStore) Workflows(params ...interface{}) ([]Workflow, error) {
	stmt := selectWorkflowsStmt
	var where []string
	for _, param := range params {
		if f, ok := param.(whereer); ok {
			where = append(where, f.where())
		}
	}

	if len(where) != 0 {
		whereFilter := strings.Join(where, ",")
		stmt = fmt.Sprintf("%s WHERE %s", selectWorkflowsStmt, whereFilter)
	}

	var workflows []Workflow
	err := store.Select(&workflows, stmt)
	if err != nil {
		return nil, errors.Wrap(err, "pgStore Workflows")
	}
	return workflows, nil
}

// whereer is for building args passed into Profiles()
type whereer interface {
	where() string
}

// sql statements
var (
	createWorkflowStmt = `INSERT INTO workflows (name) VALUES ($1) 
						 ON CONFLICT ON CONSTRAINT workflows_name_key DO NOTHING
						 RETURNING workflow_uuid;`
	selectWorkflowsStmt = `SELECT workflow_uuid, name FROM profiles`
)

//NewDB creates a Datastore
func NewDB(driver, conn string, logger kitlog.Logger) (Datastore, error) {
	switch driver {
	case "postgres":
		db, err := sqlx.Open(driver, conn)
		if err != nil {
			return nil, errors.Wrap(err, "workflow datastore")
		}
		var dbError error
		maxAttempts := 20
		for attempts := 1; attempts <= maxAttempts; attempts++ {
			dbError = db.Ping()
			if dbError == nil {
				break
			}
			logger.Log("msg", fmt.Sprintf("could not connect to postgres: %v", dbError))
			time.Sleep(time.Duration(attempts) * time.Second)
		}
		if dbError != nil {
			return nil, errors.Wrap(dbError, "workflow datastore")
		}
		migrate(db)
		return pgStore{DB: db}, nil
	default:
		return nil, errors.New("unknown driver")
	}
}

func migrate(db *sqlx.DB) {
	schema := `
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
	CREATE TABLE IF NOT EXISTS workflows (
	  workflow_uuid uuid PRIMARY KEY 
	            DEFAULT uuid_generate_v4(), 
	  name text UNIQUE NOT NULL CHECK (name <> '')
	  );

	CREATE TABLE IF NOT EXISTS profiles (
	  profile_uuid uuid PRIMARY KEY 
	            DEFAULT uuid_generate_v4(), 
	  identifier text UNIQUE NOT NULL,
	  profile_data bytea
	  );

	CREATE TABLE IF NOT EXISTS workflow_profile (
		workflow_uuid uuid REFERENCES workflows,
		profile_uuid uuid REFERENCES profiles,
		PRIMARY KEY (workflow_uuid, profile_uuid)
  	  );

	CREATE TABLE IF NOT EXISTS workflow_workflow (
		workflow_uuid uuid REFERENCES workflows,
		included_workflow_uuid uuid REFERENCES workflows(workflow_uuid),
		PRIMARY KEY (workflow_uuid, included_workflow_uuid)
  	  );`

	db.MustExec(schema)
}
