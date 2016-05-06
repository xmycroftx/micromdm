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

func TestDEPCheckin(t *testing.T) {
	client := testClient()
	defer client.teardown()
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

//helpers

//boilerplate

var (
	db        *sqlx.DB
	jsonMedia = "application/json; charset=utf-8"
	testConn  = "user=micromdm password=micromdm dbname=micromdm sslmode=disable"
)

func setup() {
	db = newDB("postgres")
}

func (c *TestClient) teardown() {
	drop := `
	DROP TABLE IF EXISTS device_workflow;
	DROP TABLE IF EXISTS devices;
	`
	db.MustExec(drop)
	c.server.Close()
}

func newDB(driver string) *sqlx.DB {
	db, err := sqlx.Open(driver, testConn)
	if err != nil {
		panic(err)
	}
	return db
}

func testServer() *httptest.Server {
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
	server := httptest.NewServer(handler)
	return server
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
	setup()
	client := &TestClient{client: http.DefaultClient}
	client.server = testServer()
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
