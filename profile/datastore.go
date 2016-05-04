package profile

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/log/levels"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgres driver
)

var (
	// ErrNoRowsModified is returned if insert didn't produce results
	ErrNoRowsModified = errors.New("db: no rows affected")
	// ErrExists is returned by AddProfile when a profile already exists
	ErrExists = errors.New("profile already exists. each profile must have a unique identifier")
	//sql stmt
	createProfileStmt = `INSERT INTO profiles (identifier, data ) VALUES ($1, $2) 
						 ON CONFLICT ON CONSTRAINT profiles_identifier_key DO NOTHING
						 RETURNING profile_uuid;`
	selectProfilesStmt = `SELECT profile_uuid, identifier, data FROM profiles`
)

// Datastore manages profiles in a database
type Datastore interface {
	AddProfile(*Profile) (*Profile, error)
	GetProfiles() ([]Profile, error)
}

type pgDatastore struct {
	info  log.Logger
	debug log.Logger
	*sqlx.DB
}

func (db pgDatastore) AddProfile(pf *Profile) (*Profile, error) {
	err := db.QueryRow(createProfileStmt, pf.PayloadIdentifier, pf.Data).Scan(&pf.UUID)
	if err == sql.ErrNoRows {
		db.debug.Log("action", "AddProfile", "profile", pf.PayloadIdentifier, "err", "exists", "status", "failure")
		return nil, ErrExists
	}
	if err != nil {
		db.info.Log("err", err, "profile", pf.PayloadIdentifier)
		return nil, errors.Wrap(err, "create profile failed")
	}
	db.debug.Log("action", "AddProfile", "profile", pf.PayloadIdentifier, "uuid", pf.UUID, "status", "success")
	return pf, nil
}

func (db pgDatastore) GetProfiles() ([]Profile, error) {
	var profiles []Profile
	err := db.Select(&profiles, selectProfilesStmt)
	if err != nil {
		db.debug.Log("action", "GetProfiles", "err", err)
		return nil, err
	}
	return profiles, nil
}

// boilerplate

type config struct {
	db      Datastore
	context context.Context
	logger  log.Logger
	debug   bool
}

func infoLogger(conf *config) log.Logger {
	return levels.New(conf.logger).Info()
}

func debugLogger(conf *config) log.Logger {
	if conf.debug {
		logger := levels.New(conf.logger).Debug()
		ctx := log.NewContext(logger).With("caller", log.DefaultCaller)
		return ctx
	}
	return log.NewNopLogger()
}

// NewDB creates a new databases connection
func NewDB(driver, conn string, options ...func(*config) error) Datastore {
	conf := &config{
		logger: log.NewNopLogger(),
	}
	defaultLogger := log.NewLogfmtLogger(os.Stderr)
	for _, option := range options {
		if err := option(conf); err != nil {
			defaultLogger.Log("err", err)
			os.Exit(1)
		}
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
		store := pgDatastore{
			info:  infoLogger(conf),
			debug: debugLogger(conf),
			DB:    db,
		}
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
	CREATE TABLE IF NOT EXISTS profiles (
	  profile_uuid uuid PRIMARY KEY 
	            DEFAULT uuid_generate_v4(), 
	  identifier text UNIQUE NOT NULL,
	  data bytea
	  );`

	db.MustExec(schema)
}
