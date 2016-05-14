package profile

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

// ErrExists is returned if a profile already exists
var ErrExists = errors.New("profile already exists. each profile must have a unique identifier")

// Datastore manages profiles in a database
type Datastore interface {
	// Add adds a profile to the datastore,
	// If a profile already exists, an error will be returned
	Add(pr *Profile) (*Profile, error)

	// Profiles can query the datastore for one or more profiles
	// and accepts one or more params as filters
	// Example: Profiles(Identifier{"com.example.id")}
	Profiles(params ...interface{}) ([]Profile, error)
}

// whereer is for building args passed into Profiles()
type whereer interface {
	where() string
}

// Identifier is a PayloadIdentifier  filter which can be passed as a param to Profiles()
type Identifier struct{ PayloadIdentifier string }

func (p Identifier) where() string {
	return fmt.Sprintf("identifier='%s'", p.PayloadIdentifier)
}

// UUID is a Profile UUID filter which can be passed as a param to Profiles()
type UUID struct{ UUID string }

func (p UUID) where() string {
	return fmt.Sprintf("profile_uuid='%s'", p.UUID)
}

type pgStore struct {
	*sqlx.DB
}

func (store pgStore) Add(prf *Profile) (*Profile, error) {
	err := store.QueryRow(addProfileStmt, prf.PayloadIdentifier, prf.ProfileData).Scan(&prf.UUID)
	if err == sql.ErrNoRows {
		return nil, ErrExists
	}
	if err != nil {
		return nil, errors.Wrap(err, "pgStore add profile")
	}
	return prf, nil
}

func (store pgStore) Profiles(params ...interface{}) ([]Profile, error) {
	stmt := selectProfilesStmt
	var where []string
	for _, param := range params {
		if f, ok := param.(whereer); ok {
			where = append(where, f.where())
		}
	}

	if len(where) != 0 {
		whereFilter := strings.Join(where, ",")
		stmt = fmt.Sprintf("%s WHERE %s", selectProfilesStmt, whereFilter)
	}

	var profiles []Profile
	err := store.Select(&profiles, stmt)
	if err != nil {
		return nil, errors.Wrap(err, "pgStore Profiles")
	}

	return profiles, nil

}

//NewDB creates a Datastore
func NewDB(driver, conn string, logger kitlog.Logger) (Datastore, error) {
	switch driver {
	case "postgres":
		db, err := sqlx.Open(driver, conn)
		if err != nil {
			return nil, errors.Wrap(err, "profile datastore")
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
			return nil, errors.Wrap(dbError, "profile datastore")
		}
		migrate(db)
		return pgStore{DB: db}, nil
	default:
		return nil, errors.New("unknown driver")
	}
}

// sql statements
var (
	addProfileStmt = `INSERT INTO profiles (identifier, profile_data) VALUES ($1, $2) 
						 ON CONFLICT ON CONSTRAINT profiles_identifier_key DO NOTHING
						 RETURNING profile_uuid;`
	selectProfilesStmt = `SELECT profile_uuid, identifier FROM profiles`
)

func migrate(db *sqlx.DB) {
	schema := `
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
	CREATE TABLE IF NOT EXISTS profiles (
	  profile_uuid uuid PRIMARY KEY 
	            DEFAULT uuid_generate_v4(), 
	  identifier text UNIQUE NOT NULL,
	  profile_data bytea
	  );`

	db.MustExec(schema)
}
