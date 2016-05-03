package workflow

import (
	"log"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/micromdm/micromdm/profile"
)

var testConn = "user=micromdm password=micromdm dbname=micromdm sslmode=disable"

func newDB(driver string) *sqlx.DB {
	db, err := sqlx.Open(driver, testConn)
	if err != nil {
		panic(err)
	}
	return db
}

func TestMain(m *testing.M) {

	db := newDB("postgres")
	setup(db)
	retCode := m.Run()
	teardown(db)

	// call with result of m.Run()
	os.Exit(retCode)
}
func setup(db *sqlx.DB) {
	teardown(db)
	schema := `
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
	CREATE TABLE IF NOT EXISTS profiles (
	  profile_uuid uuid PRIMARY KEY 
	            DEFAULT uuid_generate_v4(), 
	  identifier text UNIQUE NOT NULL
	  );`
	db.MustExec(schema)
}

func teardown(db *sqlx.DB) {
	log.Println("workflow: dropping test tables")
	drop := `
	DROP TABLE IF EXISTS workflow_workflow;
	DROP TABLE IF EXISTS workflow_profile;
	DROP TABLE IF EXISTS workflows;
	DROP TABLE IF EXISTS profiles;
	`
	db.MustExec(drop)
}

func TestNewDBConnection(t *testing.T) {
	pg := newDB("postgres")
	checkExists := `SELECT * from information_schema.tables WHERE table_name = 'workflows'`
	t.Log("workflow: testing new postgres connection")
	_ = NewDB("postgres", testConn)
	_, err := pg.Query(checkExists)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateWorkflow(t *testing.T) {
	store := NewDB("postgres", testConn)
	db := newDB("postgres")
	db.MustExec(`INSERT INTO profiles (identifier) VALUES ($1);`, "com.micromdm.test")
	db.MustExec(`INSERT INTO profiles (identifier) VALUES ($1);`, "com.micromdm.test2")
	var pf profile.Profile
	err := db.Get(&pf, "SELECT * FROM profiles WHERE identifier=$1", "com.micromdm.test")
	if err != nil {
		t.Fatal(err)
	}

	wf, err := store.CreateWorkflow("test_workflowX")
	if err != nil {
		t.Fatal(err)
	}

	wf, err = store.CreateWorkflow("test_workflowX")
	if err == nil {
		t.Fatalf("create workflow should fail on duplicate workflow, got no errors")
	}

	wf, err = store.CreateWorkflow("test_workflow")
	if err != nil {
		t.Fatal(err)
	}

	err = store.AddProfile(wf.UUID, pf.UUID)
	if err != nil {
		t.Fatal(err)
	}
	err = db.Get(&pf, "SELECT * FROM profiles WHERE identifier=$1", "com.micromdm.test2")
	if err != nil {
		t.Fatal(err)
	}
	err = store.AddProfile(wf.UUID, pf.UUID)
	if err != nil {
		t.Fatal(err)
	}
	err = store.AddProfile(wf.UUID, pf.UUID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetWorkflows(t *testing.T) {
	store := NewDB("postgres", testConn)
	workflows, err := store.GetWorkflows()
	if err != nil {
		t.Fatal(err)
	}
	if len(workflows) == 0 {
		t.Fatal("should have at least one workflow")
	}

	if len(workflows[1].Profiles) == 0 {
		t.Fatal("should have at least one profile")
	}
}

func TestRemoveProfile(t *testing.T) {
	store := NewDB("postgres", testConn)
	db := newDB("postgres")
	var pf profile.Profile
	err := db.Get(&pf, "SELECT * FROM profiles WHERE identifier=$1", "com.micromdm.test")
	if err != nil {
		t.Fatal(err)
	}

	wf, err := store.CreateWorkflow("test_workflow_two")
	if err != nil {
		t.Fatal(err)
	}

	err = store.AddProfile(wf.UUID, pf.UUID)
	if err != nil {
		t.Fatal(err)
	}

	err = store.RemoveProfile(wf.UUID, pf.UUID)
	if err != nil {
		t.Fatal(err)
	}
}
