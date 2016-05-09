package workflow

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/micromdm/micromdm/profile"
)

var (
	pg        = newDB("postgres")
	client    = NewTestClient()
	jsonMedia = "application/json; charset=utf-8"
)

func TestMain(m *testing.M) {
	db := newDB("postgres")
	setup(db)
	retCode := m.Run()
	teardown(db)
	client.teardown()
	os.Exit(retCode)
}

type TestClient struct {
	client *http.Client
	server *httptest.Server

	// Base URL for API requests.
	BaseURL *url.URL
}

func NewTestClient() *TestClient {
	db := newDB("postgres")
	setup(db)
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

// HTTP Test Code

func testWorkflow() (*Workflow, error) {
	pf := profile.Profile{
		PayloadIdentifier: "com.apple.foo",
		Data:              "foo",
	}
	createProfileStmt := `INSERT INTO profiles (identifier, data ) VALUES ($1, $2) 
						 ON CONFLICT ON CONSTRAINT profiles_identifier_key DO NOTHING
						 RETURNING profile_uuid;`
	err := pg.QueryRow(createProfileStmt, pf.PayloadIdentifier, pf.Data).Scan(&pf.UUID)
	if err != nil {
		return nil, err
	}
	return &Workflow{
		Name: "test_workflow_one",
		Profiles: []configProfile{
			configProfile{pf.UUID, pf.PayloadIdentifier},
		},
	}, nil
}

func TestHTTPCreateWorkflow(t *testing.T) {
	req, err := client.NewRequest("workflows", "", jsonMedia, "POST")
	if err != nil {
		t.Fatal(err)
	}
	wf, err := testWorkflow()
	if err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(wf)
	if err != nil {
		t.Fatal(err)
	}
	req.Body = &nopCloser{bytes.NewBuffer(body)}
	resp, err := client.Do(req, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Error("Expected", http.StatusCreated, "got", resp.StatusCode)
		io.Copy(os.Stdout, resp.Body)
	}
}

func TestHTTPListWorkflows(t *testing.T) {
	req, err := client.NewRequest("workflows", "", jsonMedia, "GET")
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("Expected", http.StatusOK, "got", resp.StatusCode)
		io.Copy(os.Stdout, resp.Body)
	}
}
