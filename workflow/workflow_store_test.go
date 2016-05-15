package workflow

import (
	"math/rand"
	"reflect"
	"testing"
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

var createWorkflowTests = []struct {
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

func TestDatastoreCreateWorkflow(t *testing.T) {
	ds := datastore(t)
	defer teardown()

	var checkErr = func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, tt := range createWorkflowTests {
		_, err := ds.CreateWorkflow(tt.in)
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
