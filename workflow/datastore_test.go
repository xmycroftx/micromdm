package workflow

import (
	"os"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
	"github.com/micromdm/micromdm/profile"
)

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
				Profiles: []profile.Profile{
					profile.Profile{
						PayloadIdentifier: "com.example.workflow.profile",
					},
				},
			},
			shouldErr: false,
		},
	}

	for _, tt := range createTests {
		wf, err := ds.Create(tt.in)
		if !tt.shouldErr {
			checkErr(err)
		}
		if tt.testErr != nil {
			if tt.testErr != err {
				t.Fatal("expected", tt.testErr, "got", err)
			}
		}
		profiles := tt.in.Profiles
		// check profiles
		// this test should not pass...
		if len(profiles) != 0 {
			if len(profiles) != len(wf.Profiles) {
				t.Log("checking profile count")
				t.Fatal("expected", len(profiles), "got", len(wf.Profiles))
			}
		}
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
