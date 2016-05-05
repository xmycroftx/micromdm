package device

import (
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
)

var (
	store     Datastore
	testConn  = "user=micromdm password=micromdm dbname=micromdm sslmode=disable"
	db        *sqlx.DB
	workflows []string
	devices   []string
)

func TestMain(m *testing.M) {
	setup()
	retCode := m.Run()
	teardown(db)
	os.Exit(retCode)
}

func setup() {
	db = newDB("postgres")
	schema := `
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
	CREATE TABLE IF NOT EXISTS workflows (
	  workflow_uuid uuid PRIMARY KEY 
	            DEFAULT uuid_generate_v4(), 
	  name text UNIQUE NOT NULL
	  );`
	db.MustExec(schema)
	store = NewDB("postgres", testConn)
	var uuid string
	err := db.QueryRow(`INSERT INTO workflows (name) VALUES ($1) RETURNING workflow_uuid;`, "com.micromdm.test").Scan(&uuid)
	if err != nil {
		panic(err)
	}
	workflows = append(workflows, uuid)
	var deviceuuid string
	err = db.QueryRow(`INSERT INTO devices (udid) VALUES ($1) RETURNING device_uuid;`, "aFooDevice").Scan(&deviceuuid)
	if err != nil {
		panic(err)
	}
	devices = append(devices, deviceuuid)
}

func newDB(driver string) *sqlx.DB {
	db, err := sqlx.Open(driver, testConn)
	if err != nil {
		panic(err)
	}
	return db
}

func teardown(db *sqlx.DB) {
	drop := `
	DROP TABLE IF EXISTS device_workflow;
	DROP TABLE IF EXISTS devices;
	DROP TABLE IF EXISTS workflows;
	`
	db.MustExec(drop)
}

func TestDBSave(t *testing.T) {
	fmt.Println(workflows)
	fmt.Println(devices)
	dev := &Device{
		UUID:     devices[0],
		Workflow: workflows[0],
	}
	err := store.Save("assign", dev)
	if err != nil {
		t.Fatal(err)
	}
	var result []struct {
		DeviceUUID   string `db:"device_uuid"`
		WorkflowUUID string `db:"workflow_uuid"`
	}
	err = db.Select(&result, `SELECT device_uuid, workflow_uuid FROM device_workflow`)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(result)

}
