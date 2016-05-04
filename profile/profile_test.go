package profile

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
)

var (
	client    = NewTestClient()
	jsonMedia = "application/json; charset=utf-8"
	testConn  = "user=micromdm password=micromdm dbname=micromdm sslmode=disable"
)

func TestMain(m *testing.M) {
	db := newDB("postgres")
	// setup(db)
	retCode := m.Run()
	teardown(db)
	// client.teardown()
	// call with result of m.Run()
	os.Exit(retCode)
}

type TestClient struct {
	client *http.Client
	server *httptest.Server

	// Base URL for API requests.
	BaseURL *url.URL
}

func NewTestClient() *TestClient {
	client := &TestClient{client: http.DefaultClient}
	client.server = newTestServer()
	client.BaseURL, _ = url.Parse(client.server.URL)
	client.BaseURL.Path = "mdm/"
	return client
}

// create testclient request
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

func (c *TestClient) teardown() {
	c.server.Close()
}

// a face io.ReadCloser for constructing request Body
type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func newTestServer() *httptest.Server {
	ctx := context.Background()
	logger := log.NewLogfmtLogger(os.Stderr)
	//
	profileDB := NewDB(
		"postgres",
		testConn,
		Logger(logger),
		Debug(),
	)

	profileSvc := NewService(DB(profileDB), Logger(logger), Debug())
	profileHandler := ServiceHandler(ctx, profileSvc)
	server := httptest.NewServer(profileHandler)
	return server
}

var store = NewDB("postgres", testConn, Logger(log.NewLogfmtLogger(os.Stderr)), Debug())

func newDB(driver string) *sqlx.DB {
	db, err := sqlx.Open(driver, testConn)
	if err != nil {
		panic(err)
	}
	return db
}

func setup(db *sqlx.DB) {
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
	drop := `
	DROP TABLE IF EXISTS profiles;
	`
	db.MustExec(drop)
}
