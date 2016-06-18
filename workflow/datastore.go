package workflow

import (
	"fmt"
	"strings"
	"time"

	kitlog "github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgres driver
	"github.com/pkg/errors"
)

// Datastore manages interactions of workflows in a database
type Datastore interface {
	// Create adds a new workflow to the datastore
	CreateWorkflow(wf *Workflow) (*Workflow, error)

	// Workflows can query the datastore for workflows
	// Workflows accepts one or more params as filters
	// to narrow down the number of results
	Workflows(params ...interface{}) ([]Workflow, error)

	// UpdateWorkflow saves changes to a workflow in the datastore
	UpdateWorkflow(wf *Workflow) (*Workflow, error)

	// CreateProfile adds a new profile to the datastore,
	// If a profile already exists, an error will be returned
	CreateProfile(p *Profile) (*Profile, error)

	// DeleteProfile removes a profile from the datastore
	DeleteProfile(pr *Profile) error

	// Profiles can query the datastore for profiles
	// Profiles accepts one or more params as filters
	// to narrow down the number of results
	Profiles(params ...interface{}) ([]Profile, error)
}

type pgStore struct {
	*sqlx.DB
}

// whereer is for building args passed into a method which finds resources
type whereer interface {
	where() string
}

// add WHERE clause from params
func addWhereFilters(stmt string, params ...interface{}) string {
	var where []string
	for _, param := range params {
		if f, ok := param.(whereer); ok {
			where = append(where, f.where())
		}
	}

	if len(where) != 0 {
		whereFilter := strings.Join(where, ",")
		stmt = fmt.Sprintf("%s WHERE %s", stmt, whereFilter)
	}
	return stmt
}

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
		return pgStore{DB: db}, nil
	default:
		return nil, errors.New("unknown driver")
	}
}
