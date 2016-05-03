// Package workflow manages device workflows
package workflow

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgres driver
	"github.com/micromdm/micromdm/profile"
	"golang.org/x/net/context"
)

// ErrNoRowsModified is returned if insert didn't produce results
var ErrNoRowsModified = errors.New("db: no rows affected")

var (
	createWorkflowStmt = `INSERT INTO workflows (name) VALUES ($1) 
						 ON CONFLICT ON CONSTRAINT workflows_name_key DO NOTHING
						 RETURNING workflow_uuid;`

	getWorkflowByNameStmt = `SELECT workflow_uuid, name FROM workflows WHERE name=$1`
	selectWorkflowsStmt   = `SELECT workflow_uuid, name FROM workflows`
	addProfileStmt        = `INSERT INTO workflow_profile (workflow_uuid, profile_uuid) VALUES ($1, $2)
								  ON CONFLICT ON CONSTRAINT workflow_profile_pkey DO NOTHING;`
	removeProfileStmt          = `DELETE FROM workflow_profile WHERE workflow_uuid=$1 AND profile_uuid=$2;`
	getProfilesForWorkflowStmt = `SELECT profiles.profile_uuid,identifier FROM profiles 
								  LEFT JOIN workflow_profile 
								  ON workflow_profile.profile_uuid = profiles.profile_uuid 
								  WHERE workflow_profile.workflow_uuid=$1`
)

type application struct {
	ManagementFlags int
	ManifestURL     string
}

// Workflow is a device workflow
type Workflow struct {
	UUID     string `db:"workflow_uuid"`
	Name     string `db:"name"`
	Profiles []profile.Profile
	// Applications      []application
	// IncludedWorkflows []Workflow
}

// Datastore manages interactions of workflows in a database
type Datastore interface {
	CreateWorkflow(string) (*Workflow, error)
	AddProfile(string, string) error
	RemoveProfile(string, string) error
	GetWorkflows() ([]Workflow, error)
}

type pgDatastore struct {
	*sqlx.DB
}

// CreateWorkflow creates a new workflow
func (db pgDatastore) CreateWorkflow(name string) (*Workflow, error) {
	workflow := &Workflow{Name: name}
	err := db.QueryRow(createWorkflowStmt, name).Scan(&workflow.UUID)
	if err == sql.ErrNoRows {
		return nil, ErrNoRowsModified
	}
	if err != nil {
		return nil, err
	}
	return workflow, nil
}

// GetWorkflows returns all workflows in the database
func (db pgDatastore) GetWorkflows() ([]Workflow, error) {
	var workflows []Workflow
	err := db.Select(&workflows, selectWorkflowsStmt)
	if err != nil {
		return nil, err
	}
	var withProfiles []Workflow
	for _, wf := range workflows {
		profiles, err := db.getProfilesForWorkflow(wf.UUID)
		if err != nil {
			return nil, err
		}
		wf.Profiles = profiles
		withProfiles = append(withProfiles, wf)
	}
	return withProfiles, nil
}

// AddProfile adds a profile to a workflow
func (db pgDatastore) AddProfile(wfUUID, pfUUID string) error {
	result, err := db.Exec(
		addProfileStmt,
		wfUUID,
		pfUUID,
	)
	if err != nil {
		return err
	}
	if _, err := result.RowsAffected(); err != nil {
		return err
	}
	return nil
}

// RemoveProfile removes a profile wrom a workflow
func (db pgDatastore) RemoveProfile(wfUUID, pfUUID string) error {
	result, err := db.Exec(
		removeProfileStmt,
		wfUUID,
		pfUUID,
	)
	if err != nil {
		return err
	}
	if _, err := result.RowsAffected(); err != nil {
		return err
	}
	return nil
}

func (db pgDatastore) getProfilesForWorkflow(workflowUUID string) ([]profile.Profile, error) {
	var profiles []profile.Profile
	err := db.Select(&profiles, getProfilesForWorkflowStmt, workflowUUID)
	if err != nil {
		return nil, err
	}
	return profiles, nil
}

// Logger adds a logger to the database config
func Logger(logger log.Logger) func(*config) error {
	return func(c *config) error {
		c.logger = logger
		return nil
	}
}

type config struct {
	context context.Context
	logger  log.Logger
}

// NewDB creates a new databases connection
func NewDB(driver, conn string, options ...func(*config) error) Datastore {
	conf := &config{}
	defaultLogger := log.NewLogfmtLogger(os.Stderr)
	for _, option := range options {
		if err := option(conf); err != nil {
			defaultLogger.Log("err", err)
			os.Exit(1)
		}
	}
	if conf.logger == nil {
		conf.logger = log.NewNopLogger()
	}
	switch driver {
	case "postgres":
		db, err := sqlx.Open(driver, conn)
		if err != nil {
			conf.logger.Log("err", err)
			os.Exit(1)
		}
		var dbError error
		maxAttempts := 20
		for attempts := 1; attempts <= maxAttempts; attempts++ {
			dbError = db.Ping()
			if dbError == nil {
				break
			}
			conf.logger.Log("msg", fmt.Sprintf("could not connect to postgres: %v", dbError))
			time.Sleep(time.Duration(attempts) * time.Second)
		}
		if dbError != nil {
			conf.logger.Log("err", dbError)
			os.Exit(1)
		}
		migrate(db)
		// TODO: configurable with default
		db.SetMaxOpenConns(5)
		store := pgDatastore{db}
		return store
	default:
		conf.logger.Log("err", "unknown driver")
		os.Exit(1)
		return nil
	}
}

func migrate(db *sqlx.DB) {
	schema := `
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
	CREATE TABLE IF NOT EXISTS workflows (
	  workflow_uuid uuid PRIMARY KEY 
	            DEFAULT uuid_generate_v4(), 
	  name text UNIQUE NOT NULL
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
