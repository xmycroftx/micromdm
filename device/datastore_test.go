package device

import (
	"os"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
)

func TestSaveTokenUpdate(t *testing.T) {
	ds := datastore(t)
	defer teardown()
	devices := addTestDevices(t, ds)

	for _, d := range devices {
		dev := &Device{
			UUID: d.UUID,
			AwaitingConfiguration: true,
			PushMagic:             "some-magic-token",
			Token:                 "some-mdm-token",
			Enrolled:              true,
		}
		err := ds.Save("tokenUpdate", dev)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestNewDB(t *testing.T) {
	defer teardown()
	logger := log.NewLogfmtLogger(os.Stderr)
	_, err := NewDB("postgres", testConn, logger)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRetrieveDevices(t *testing.T) {
	ds := datastore(t)
	defer teardown()
	devices := addTestDevices(t, ds)

	_, err := ds.Devices()
	if err != nil {
		t.Fatal(err)
	}

	dev0 := devices[0]
	returned, err := ds.Devices(UUID{dev0.UUID})
	if err != nil {
		t.Fatal(err)
	}
	if returned[0].SerialNumber != dev0.SerialNumber {
		t.Fatal("expected", dev0.SerialNumber, "got", returned[0].SerialNumber)
	}
}

func addTestDevices(t *testing.T, ds Datastore) []Device {
	now := time.Now()
	var devicetests = []struct {
		in Device
	}{
		{
			Device{
				SerialNumber: "DEADBEEF123A",
				Model:        "Macbook",
				Description:  "It's a laptop",
				Color:        "red",
			},
		},
		{
			Device{
				SerialNumber: "DEADBEEF123A",
				Model:        "Macbook",
				Description:  "It's a laptop",
				Color:        "red",
			},
		},
		{
			Device{
				SerialNumber: "DEADBEEF123B",
				Model:        "Macbook",
				Description:  "It's a laptop",
				Color:        "blue",
			},
		},
		{
			Device{
				SerialNumber:         "DEADBEEF123C",
				Model:                "iPad",
				Description:          "It's a tablet",
				Color:                "pink",
				AssetTag:             "foo",
				DEPProfileAssignTime: &now,
			},
		},
	}

	var devices []Device

	for _, tt := range devicetests {
		uuid, err := ds.New("fetch", &tt.in)
		if err != nil {
			t.Log("failed at", tt.in.SerialNumber)
			t.Fatal(err)
		}
		if len(uuid) != 36 {
			t.Errorf("newdevice fetch: expected uuid got %q", uuid)
		}

		d := tt.in
		d.UUID = uuid
		devices = append(devices, d)
	}
	return devices
}
func TestInsertFetch(t *testing.T) {
	ds := datastore(t)
	defer teardown()

	now := time.Now()
	var devicetests = []struct {
		in Device
	}{
		{
			Device{
				SerialNumber: "DEADBEEF123A",
				Model:        "Macbook",
				Description:  "It's a laptop",
				Color:        "red",
			},
		},
		{
			Device{
				SerialNumber: "DEADBEEF123A",
				Model:        "Macbook",
				Description:  "It's a laptop",
				Color:        "red",
			},
		},
		{
			Device{
				SerialNumber: "DEADBEEF123B",
				Model:        "Macbook",
				Description:  "It's a laptop",
				Color:        "blue",
			},
		},
		{
			Device{
				SerialNumber:         "DEADBEEF123C",
				Model:                "iPad",
				Description:          "It's a tablet",
				Color:                "pink",
				AssetTag:             "foo",
				DEPProfileAssignTime: &now,
			},
		},
	}

	for _, tt := range devicetests {
		uuid, err := ds.New("fetch", &tt.in)
		if err != nil {
			t.Log("failed at", tt.in.SerialNumber)
			t.Fatal(err)
		}
		if len(uuid) != 36 {
			t.Errorf("newdevice fetch: expected uuid got %q", uuid)
		}
	}
}

func TestInsertAuthenticate(t *testing.T) {
	ds := datastore(t)
	defer teardown()

	var now = time.Now()
	var devicetests = []struct {
		in Device
	}{
		{
			Device{
				SerialNumber: "DEADBEEF123A",
				Model:        "Macbook",
				Description:  "It's a laptop",
				Color:        "red",
			},
		},
		{
			Device{
				SerialNumber: "DEADBEEF123A",
				Model:        "Macbook",
				Description:  "It's a laptop",
				Color:        "red",
			},
		},
		{
			Device{
				SerialNumber: "DEADBEEF123B",
				Model:        "Macbook",
				Description:  "It's a laptop",
				Color:        "blue",
			},
		},
		{
			Device{
				UDID:         "581ddbee-7742-4472-aadd-6d2ad35c4470",
				SerialNumber: "DEADBEEF123B",
				Model:        "Macbook",
				Description:  "It's a laptop",
				Color:        "blue",
			},
		},
		{
			Device{
				UDID:         "581ddbee-7742-4472-aadd-6d2ad35c4470",
				SerialNumber: "DEADBEEF123B",
				Model:        "Macbook",
				Description:  "It's a laptop",
				Color:        "blue",
			},
		},
		{
			Device{
				SerialNumber:         "DEADBEEF123C",
				Model:                "iPad",
				Description:          "It's a tablet",
				Color:                "pink",
				AssetTag:             "foo",
				DEPProfileAssignTime: &now,
			},
		},
	}

	for _, tt := range devicetests {
		uuid, err := ds.New("authenticate", &tt.in)
		if err != nil {
			t.Log("failed at", tt.in.SerialNumber)
			t.Fatal(err)
		}
		if len(uuid) != 36 {
			t.Errorf("newdevice authenticate: expected uuid got %q", uuid)
		}
	}
}

func TestGetDeviceByUDID(t *testing.T) {
	ds := datastore(t)
	defer teardown()
	var now = time.Now()
	var devicetests = []struct {
		in Device
	}{
		{
			Device{
				UDID:         "581ddbee-7742-4472-aadd-6d2ad35c4471",
				SerialNumber: "DEADBEEF123A",
				Model:        "Macbook",
				Description:  "It's a laptop",
				Color:        "red",
			},
		},
		{
			Device{
				UDID:         "581ddbee-7742-4472-aadd-6d2ad35c4471",
				SerialNumber: "DEADBEEF123A",
				Model:        "Macbook",
				Description:  "It's a laptop",
				Color:        "red",
			},
		},
		{
			Device{
				UDID:         "581ddbee-7742-4472-aadd-6d2ad35c4470",
				SerialNumber: "DEADBEEF123B",
				Model:        "Macbook",
				Description:  "It's a laptop",
				Color:        "blue",
			},
		},
		{
			Device{
				UDID:         "581ddbee-7742-4472-aadd-6d2ad35c4470",
				SerialNumber: "DEADBEEF123B",
				Model:        "Macbook",
				Description:  "It's a laptop",
				Color:        "blue",
			},
		},
		{
			Device{
				UDID:         "581ddbee-7742-4472-aadd-6d2ad35c4470",
				SerialNumber: "DEADBEEF123B",
				Model:        "Macbook",
				Description:  "It's a laptop",
				Color:        "blue",
			},
		},
		{
			Device{
				UDID:                 "581ddbee-7742-4472-aadd-6d2ad35c4472",
				SerialNumber:         "DEADBEEF123C",
				Model:                "iPad",
				Description:          "It's a tablet",
				Color:                "pink",
				AssetTag:             "foo",
				DEPProfileAssignTime: &now,
			},
		},
	}

	for _, tt := range devicetests {
		uuid, err := ds.New("authenticate", &tt.in)
		if err != nil {
			t.Log("failed at", tt.in.SerialNumber)
			t.Fatal(err)
		}
		if len(uuid) != 36 {
			t.Errorf("newdevice get device by udid: expected uuid got %q", uuid)
		}
		d, err := ds.GetDeviceByUDID(tt.in.UDID, "device_uuid", "udid", "serial_number")
		if err != nil {
			t.Log("get failed at", tt.in.SerialNumber)
			t.Fatal(err)
		}
		if d.SerialNumber != tt.in.SerialNumber {
			t.Errorf("get device by udid: expected %q got %q", tt.in.SerialNumber, d.SerialNumber)
		}
	}

}

var (
	testConn = "user=micromdm password=micromdm dbname=micromdm sslmode=disable"
)

func datastore(t *testing.T) Datastore {
	//setup()
	logger := log.NewLogfmtLogger(os.Stderr)
	ds, err := NewDB("postgres", testConn, logger)
	if err != nil {
		t.Fatal(err)
	}
	return ds
}

func setup() {
	db, err := sqlx.Open("postgres", testConn)
	if err != nil {
		panic(err)
	}
	migrate(db)
	defer db.Close()
}

func teardown() {
	db, err := sqlx.Open("postgres", testConn)
	if err != nil {
		panic(err)
	}

	drop := `
	DROP TABLE IF EXISTS device_workflow;
	DROP TABLE IF EXISTS devices;
	DROP INDEX IF EXISTS devices.serial_idx;
	DROP INDEX IF EXISTS devices.udid_idx;
	`
	db.MustExec(drop)
	defer db.Close()
}
