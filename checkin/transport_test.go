package checkin

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
	"github.com/micromdm/micromdm/device"
	"golang.org/x/net/context"
)

// test DEP enrollment request
func TestEnroll(t *testing.T) {
	req, err := client.NewRequest("checkin", "", jsonMedia, "POST")
	if err != nil {
		t.Fatalf("could not create POST request to /mdm/checkin: \n %q", err)
	}
	file := filepath.Join("test-fixtures", "dat1")
	data, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	req.Body = &nopCloser{bytes.NewBuffer(data)}
	resp, err := client.Do(req, nil)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		io.Copy(os.Stdout, resp.Body)
		t.Fatal("expected", http.StatusOK, "got", resp.Status)
	}
}

func TestAuthenticate(t *testing.T) {
	req, err := client.NewRequest("checkin", "", jsonMedia, "PUT")
	if err != nil {
		t.Fatalf("could not create PUT request to /mdm/checkin: \n %q", err)
	}
	req.Body = &nopCloser{bytes.NewBuffer(authRequest)}
	resp, err := client.Do(req, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		io.Copy(os.Stdout, resp.Body)
		t.Fatal("expected", http.StatusOK, "got", resp.Status)
	}
}

func TestTokenUpdate(t *testing.T) {
	req, err := client.NewRequest("checkin", "", jsonMedia, "PUT")
	if err != nil {
		t.Fatalf("could not create PUT request to /mdm/checkin: \n %q", err)
	}
	req.Body = &nopCloser{bytes.NewBuffer(tokenUpdateRequest)}
	resp, err := client.Do(req, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		io.Copy(os.Stdout, resp.Body)
		t.Fatal("expected", http.StatusOK, "got", resp.Status)
	}
}

//helpers
var authRequest = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
        <key>BuildVersion</key>
        <string>15E65</string>
        <key>Challenge</key>
        <data>
        YXBwbGU=
        </data>
        <key>DeviceName</key>
        <string>Mac</string>
        <key>MessageType</key>
        <string>Authenticate</string>
        <key>Model</key>
        <string>VMware7,1</string>
        <key>ModelName</key>
        <string>Apple device</string>
        <key>OSVersion</key>
        <string>10.11.4</string>
        <key>ProductName</key>
        <string>VMware7,1</string>
        <key>SerialNumber</key>
        <string>DEADBEEF123K</string>
        <key>Topic</key>
        <string>com.apple.mgmt.External.db27e349-c6fb-4927-a393-129938355b82</string>
        <key>UDID</key>
        <string>564D4831-42F8-3657-5E37-E433037134D0</string>
</dict>
</plist>`)

var tokenUpdateRequest = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
        <key>AwaitingConfiguration</key>
        <true/>
        <key>MessageType</key>
        <string>TokenUpdate</string>
        <key>PushMagic</key>
        <string>326BFD45-7E76-48A4-BEC6-538520C1ECD1</string>
        <key>Token</key>
        <data>
        dYty6tVxXjNtTD8c4bsWvqGOm5Y3563KiOPQqqjzSHY=
        </data>
        <key>Topic</key>
        <string>com.apple.mgmt.External.db27e349-c6fb-4927-a393-129938355b82</string>
        <key>UDID</key>
        <string>564D4831-42F8-3657-5E37-E433037134D0</string>
</dict>
</plist>`)

//boilerplate

var (
	client    *TestClient
	server    *httptest.Server
	db        *sqlx.DB
	jsonMedia = "application/json; charset=utf-8"
	testConn  = "user=micromdm password=micromdm dbname=micromdm sslmode=disable"
)

func TestMain(m *testing.M) {
	setup()
	retCode := m.Run()
	teardown()
	os.Exit(retCode)
}

func setup() {
	testServer()
	client = testClient()
	db = newDB("postgres")
}

func teardown() {
	drop := `
	DROP TABLE IF EXISTS device_workflow;
	DROP TABLE IF EXISTS devices;
	`
	db.MustExec(drop)
}

func newDB(driver string) *sqlx.DB {
	db, err := sqlx.Open(driver, testConn)
	if err != nil {
		panic(err)
	}
	return db
}

func testServer() {
	logger := log.NewLogfmtLogger(os.Stderr)
	deviceDB := device.NewDB(
		"postgres",
		testConn,
		device.Logger(logger),
		device.EnrollProfile("test-fixtures/dat1"),
	)
	ctx := context.Background()
	checkinSvc := NewCheckinService(
		Datastore(deviceDB),
		Logger(logger),
		// Push(*flPushCert, *flPushPass),
	)
	handler := ServiceHandler(ctx, checkinSvc)
	server = httptest.NewServer(handler)
}

// a face io.ReadCloser for constructing request Body
type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

type TestClient struct {
	client *http.Client
	server *httptest.Server

	// Base URL for API requests.
	BaseURL *url.URL
}

func testClient() *TestClient {
	client := &TestClient{client: http.DefaultClient}
	client.server = server
	client.BaseURL, _ = url.Parse(client.server.URL)
	client.BaseURL.Path = "mdm/"
	return client
}

func (c *TestClient) NewRequest(endpoint, resource, mediaType, method string) (*http.Request, error) {
	var urlStr string
	if resource != "" {
		urlStr = endpoint + "/" + resource
	} else {
		urlStr = endpoint
	}
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	u := c.BaseURL.ResolveReference(rel)
	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, err
	}
	return req, nil

}

// run the request
func (c *TestClient) Do(req *http.Request, into interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
