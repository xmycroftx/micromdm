package device

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgres driver
)

// ErrNoRowsModified is returned if insert didn't produce results
var ErrNoRowsModified = errors.New("DB: No rows affected")

// Device represents an iOS or OS X Computer
type Device struct {
	// Primary key is UUID
	UUID         string  `json:"uuid" db:"device_uuid"`
	UDID         string  `json:"udid"`
	SerialNumber *string `json:"serial_number,omitempty" db:"serial_number,omitempty"`
	OSVersion    *string `json:"os_version,omitempty" db:"os_version,omitempty"`
	BuildVersion *string `json:"build_version,omitempty" db:"build_version,omitempty"`
	ProductName  *string `json:"product_name,omitempty" db:"product_name,omitempty"`
	IMEI         *string `json:"imei,omitempty" db:"imei,omitempty"`
	MEID         *string `json:"meid,omitempty" db:"meid,omitempty"`
	//Apple MDM Protocol Topic
	MDMTopic              *string `json:"mdm_topic,omitempty" db:"apple_mdm_topic,omitempty"`
	PushMagic             *string `json:"push_magic,omitempty" db:"apple_push_magic,omitempty"`
	AwaitingConfiguration *bool   `json:"awaiting_configuration,omitempty" db:"awaiting_configuration,omitempty"`
	Token                 *string `json:"token,omitempty" db:"apple_mdm_token,omitempty"`
	UnlockToken           *string `json:"unlock_token,omitempty" db:"unlock_token,omitempty"`
	Enrolled              *bool   `json:"enrolled,omitempty" db:"mdm_enrolled,omitempty"`
	Workflow              string  `json:"workflow,omitempty" db:"workflow_uuid"`
}

// Profile is an Enrollment profile.
// For now we just have one.
type Profile []byte

// Datastore manages interactions of devices in a database
type Datastore interface {
	AddDevice(*Device) error
	GetDeviceByUDID(udid string) (*Device, error)
	SaveDevice(*Device) error
	// RemoveDevice() error
	// AllDevices() DeviceList,error
	GetProfileForDevice(udid string) (*Profile, error)
	Save(string, *Device) error
}

type config struct {
	context       context.Context
	logger        log.Logger
	enrollProfile string
}

// NewDB creates a new databases connection
func NewDB(driver, conn string, options ...func(*config) error) Datastore {
	conf := &config{}
	defaultLogger := log.NewLogfmtLogger(os.Stderr)
	for _, option := range options {
		if err := option(conf); err != nil {
			defaultLogger.Log("err", err)
			os.Exit(1)
		}
	}
	if conf.logger == nil {
		conf.logger = defaultLogger
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
		store := pgDatastore{db, conf.enrollProfile}
		return store
	default:
		conf.logger.Log("err", "unknown driver")
		os.Exit(1)
		return nil
	}
}

// Logger adds a logger to the database config
func Logger(logger log.Logger) func(*config) error {
	return func(c *config) error {
		c.logger = logger
		return nil
	}
}

// EnrollProfile sets the path for the enrollment profile
// another temp hack
func EnrollProfile(path string) func(*config) error {
	return func(c *config) error {
		c.enrollProfile = path
		return nil
	}
}

// datastore implementation for postgres
type pgDatastore struct {
	*sqlx.DB
	enrollProfile string
}

func (db pgDatastore) AddDevice(dev *Device) error {
	upsert := `INSERT INTO devices 
               (udid, apple_mdm_topic, os_version, build_version, product_name, serial_number, imei, meid)
               VALUES ($1,$2,$3,$4,$5,$6,$7,$8) 
               ON CONFLICT ON CONSTRAINT devices_udid_key 
               DO UPDATE SET 
                   apple_mdm_topic=$2,
                   os_version=$3,
                   build_version=$4,
                   product_name=$5,
                   serial_number=$6,
                   imei=$7,
                   meid=$8;`
	result, err := db.Exec(
		upsert,
		dev.UDID,
		dev.MDMTopic,
		dev.OSVersion,
		dev.BuildVersion,
		dev.ProductName,
		dev.SerialNumber,
		dev.IMEI,
		dev.MEID,
	)
	if err != nil {
		return err
	}
	if res, _ := result.RowsAffected(); res == 0 {
		return ErrNoRowsModified
	}
	return nil
}

func (db pgDatastore) GetDeviceByUDID(udid string) (*Device, error) {
	var device Device
	query := `SELECT * FROM devices WHERE udid=$1 LIMIT 1`
	return &device, sqlx.Get(db, &device, query, udid)
}

// SaveDevice updates a device with the latest changes
func (db pgDatastore) SaveDevice(dev *Device) error {
	update := `UPDATE devices SET
	awaiting_configuration=$2,
	apple_push_magic=$3,
	apple_mdm_token=$4,
	mdm_enrolled=$5
	WHERE device_uuid=$1`
	result, err := db.Exec(
		update,
		dev.UUID,
		dev.AwaitingConfiguration,
		dev.PushMagic,
		dev.Token,
		dev.Enrolled,
	)
	if err != nil {
		return err
	}
	if res, _ := result.RowsAffected(); res == 0 {
		return ErrNoRowsModified
	}
	return nil
}

func (db pgDatastore) Save(msg string, dev *Device) error {
	var stmt string
	switch msg {
	case "assign":
		stmt = `INSERT INTO device_workflow VALUES (:device_uuid, :workflow_uuid);`
	}
	_, err := db.NamedExec(stmt, dev)
	if err != nil {
		return err
	}
	return nil
}

// GetProfileForDevice returns an enrollment profile for a specific device
// For now there's a single profile stored on disk
func (db pgDatastore) GetProfileForDevice(uuid string) (*Profile, error) {
	data, err := ioutil.ReadFile(db.enrollProfile)
	if err != nil {
		return nil, err
	}
	profile := Profile(data)
	return &profile, nil
}

func migrate(db *sqlx.DB) {
	schema := `
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
	CREATE TABLE IF NOT EXISTS devices (
	  device_uuid uuid PRIMARY KEY 
	            DEFAULT uuid_generate_v4(), 
	  udid text UNIQUE NOT NULL,
	  serial_number text,
	  os_version text,
	  build_version text,
	  product_name text,
	  imei text,
	  meid text,
	  apple_mdm_token text,
	  apple_mdm_topic text,
	  apple_push_magic text,
	  mdm_enrolled boolean,
	  awaiting_configuration boolean
	  );

	CREATE TABLE IF NOT EXISTS device_workflow (
		device_uuid uuid REFERENCES devices,
		workflow_uuid uuid REFERENCES workflows,
		PRIMARY KEY (device_uuid, workflow_uuid)
  	  );`
	db.MustExec(schema)
}
