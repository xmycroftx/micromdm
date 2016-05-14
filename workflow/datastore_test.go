package workflow

import (
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
)

func TestDatastoreWorkflows(t *testing.T) {
	// ds := datastore(t)
	defer teardown()
}

func (wf Workflow) Generate(rand *rand.Rand, size int) reflect.Value {
	name := randomString(16)
	randomWorkflow := Workflow{
		Name: name,
	}

	return reflect.ValueOf(randomWorkflow)
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

func TestNewDB(t *testing.T) {
	_ = datastore(t)
	defer teardown()
}

func TestDatastoreCreate(t *testing.T) {
	ds := datastore(t)
	defer teardown()

	var checkErr = func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}

	var createTests = []struct {
		in        *Workflow
		shouldErr bool
		testErr   error
	}{
		{
			in:        &Workflow{},
			shouldErr: true,
		},
		{
			in: &Workflow{
				Name: "exampleWorkflow",
			},
			shouldErr: false,
		},
		{
			in: &Workflow{
				Name: "exampleWorkflow",
			},
			shouldErr: true,
			testErr:   ErrExists,
		},
		{
			in: &Workflow{
				Name: "exampleWorkflowWithProfiles",
				Profiles: []Profile{
					Profile{UUID: "c7616875-df2d-4fe5-9c1e-0cb36c1ede8a"},
				},
			},
			shouldErr: true,
		},
	}

	for _, tt := range createTests {
		_, err := ds.Create(tt.in)
		if !tt.shouldErr {
			checkErr(err)
		}
		if tt.testErr != nil {
			if tt.testErr != err {
				t.Fatal("expected", tt.testErr, "got", err)
			}
		}
		// check profiles
	}
}

func datastore(t *testing.T) Datastore {
	logger := log.NewLogfmtLogger(os.Stderr)
	ds, err := NewDB("postgres", testConn, logger)
	if err != nil {
		t.Fatal(err)
	}
	return ds
}

var (
	testConn = "user=micromdm password=micromdm dbname=micromdm sslmode=disable"
)

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
