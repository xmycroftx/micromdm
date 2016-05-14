package workflow

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// sql statements
var (
	addProfileStmt = `INSERT INTO profiles (identifier, profile_data) VALUES ($1, $2) 
						 ON CONFLICT ON CONSTRAINT profiles_identifier_key DO NOTHING
						 RETURNING profile_uuid;`
	selectProfilesStmt = `SELECT profile_uuid, identifier FROM profiles`
)

func (store pgStore) CreateProfile(prf *Profile) (*Profile, error) {
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
