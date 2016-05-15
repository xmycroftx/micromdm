package workflow

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
)

// sql statements
var (
	addProfileStmt = `INSERT INTO profiles 
					  (payload_identifier, profile_data) VALUES ($1, $2) 
					  ON CONFLICT ON CONSTRAINT profiles_payload_identifier_key DO NOTHING
					  RETURNING profile_uuid;`

	selectProfilesStmt = `SELECT profile_uuid, payload_identifier, profile_data FROM profiles`
	deleteProfileStmt  = `DELETE FROM profiles`
)

// ProfileUUID is a filter we can add as a parameter to narrow down the list of returned results
type ProfileUUID struct {
	UUID string
}

func (p ProfileUUID) where() string {
	return fmt.Sprintf("profile_uuid = '%s'", p.UUID)
}

// PayloadIdentifier is a filter we can add as a parameter to narrow down the list of returned results
type PayloadIdentifier struct {
	PayloadIdentifier string
}

func (p PayloadIdentifier) where() string {
	return fmt.Sprintf("payload_identifier = '%s'", p.PayloadIdentifier)
}

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

func (store pgStore) DeleteProfile(p *Profile) error {
	stmt := deleteProfileStmt
	if p.UUID == "" || p.PayloadIdentifier == "" {
		return nil // just don't do anything
	}
	if p.UUID != "" {
		stmt = addWhereFilters(stmt, ProfileUUID{p.UUID})
	}
	_, err := store.Exec(stmt)
	if err != nil {
		return errors.Wrap(err, "delete profile")
	}
	return nil
}

func (store pgStore) Profiles(params ...interface{}) ([]Profile, error) {
	stmt := selectProfilesStmt
	stmt = addWhereFilters(stmt, params...)

	var profiles []Profile
	err := store.Select(&profiles, stmt)
	if err != nil {
		return nil, errors.Wrap(err, "pgStore Profiles")
	}

	return profiles, nil

}
