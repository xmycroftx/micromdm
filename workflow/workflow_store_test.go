package workflow

import (
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

func TestUpdateWorkflow(t *testing.T) {
	ds := datastore(t)
	defer teardown()
	testProfiles := addTestProfiles(t, ds, 10)
	testWorkflows := addTestWorkflows(t, ds, 5)
	addProfileWorkflow := testWorkflows[0]
	addProfileWorkflow.Profiles = append(addProfileWorkflow.Profiles, testProfiles[0])
	_, err := ds.UpdateWorkflow(&addProfileWorkflow)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := ds.Workflows(WrkflowUUID{addProfileWorkflow.UUID})
	if err != nil {
		t.Fatal(err)
	}
	if !updated[0].HasProfile(testProfiles[0].PayloadIdentifier) {
		t.Fatal("expected workflow to have added profiles")
	}

	remProfileWf := updated[0]
	remProfileWf.Profiles = []Profile{}
	_, err = ds.UpdateWorkflow(&remProfileWf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRetrieveWorkflows(t *testing.T) {
	ds := datastore(t)
	defer teardown()
	testWorkflows := addTestWorkflows(t, ds, 5)
	workflows, err := ds.Workflows()
	if err != nil {
		t.Fatal(err)
	}

	if len(workflows) != 5 {
		t.Error("expected", 5, "got", len(workflows))
	}

	for _, p := range testWorkflows {
		byUUID, err := ds.Workflows(WrkflowUUID{p.UUID})
		if err != nil {
			t.Fatal(err)
		}
		if len(byUUID) != 1 {
			t.Log("filtering by UUID should only return 1 result")
			t.Fatal("expected", 1, "got", len(byUUID))
		}

		uuid := byUUID[0].UUID
		if p.UUID != uuid {
			t.Log("result should have the same UUID as the one in the query")
			t.Fatal("expected", p.UUID, "got", uuid)

		}
	}

	// test with error
	badUUIDQuery := ProfileUUID{"bad_uuid"}
	_, err = ds.Workflows(badUUIDQuery)
	if err == nil {
		t.Fatal("expected an error but got nil")

	}
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

func TestDatastoreCreateWorkflowWithProfiles(t *testing.T) {
	ds := datastore(t)
	defer teardown()
	testProfiles := addTestProfiles(t, ds, 5)

	wf := &Workflow{
		Name:     "has_profiles",
		Profiles: testProfiles,
	}
	wf, err := ds.CreateWorkflow(wf)
	if err != nil {
		t.Fatal(err)
	}
	if wf.UUID == "" {
		t.Fatal("expected nonempty uuid result")
	}

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
	}
}

func addTestWorkflows(t *testing.T, ds Datastore, numWorkflows int) []Workflow {
	var workflows []Workflow
	for i := 0; i < numWorkflows; i++ {
		input := randomWorkflow()
		newWorkflow, err := ds.CreateWorkflow(&input)
		if err != nil {
			t.Fatal(err)
		}
		workflows = append(workflows, *newWorkflow)
	}
	return workflows

}

func randomWorkflow() Workflow {
	vrf, ok := quick.Value(reflect.TypeOf(Workflow{}), rand.New(rand.NewSource(1)))
	if !ok {
		panic("randomProfile: no value")
	}
	if f, ok := vrf.Interface().(Workflow); ok {
		return f
	}
	return Workflow{}
}

func (wf Workflow) Generate(rand *rand.Rand, size int) reflect.Value {
	name := randomString(16)
	randomWorkflow := Workflow{
		Name: name,
	}

	return reflect.ValueOf(randomWorkflow)
}
