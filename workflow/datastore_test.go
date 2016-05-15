package workflow

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
)

// Test that a new datastore is succesfuly created and destroyed.
func TestNewDB(t *testing.T) {
	_ = datastore(t)
	defer teardown()
}

var (
	testConn = "user=micromdm password=micromdm dbname=micromdm sslmode=disable"
)

func datastore(t *testing.T) Datastore {
	logger := log.NewLogfmtLogger(os.Stderr)
	ds, err := NewDB("postgres", testConn, logger)
	if err != nil {
		t.Fatal(err)
	}
	return ds
}

func teardown() {
	db, err := sqlx.Open("postgres", testConn)
	if err != nil {
		panic(err)
	}

	drop := `
	DROP TABLE IF EXISTS workflow_profile;
	DROP TABLE IF EXISTS workflow_workflow;
	DROP TABLE IF EXISTS profiles;
	DROP TABLE IF EXISTS workflows;
	`
	db.MustExec(drop)
	defer db.Close()
}

func randomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
