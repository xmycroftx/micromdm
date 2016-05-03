package workflow

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/go-kit/kit/log"
	"golang.org/x/net/context"
)

var (
	client    = NewTestClient()
	jsonMedia = "application/json; charset=utf-8"
)

func newTestServer() *httptest.Server {
	ctx := context.Background()
	logger := log.NewLogfmtLogger(os.Stderr)
	//
	workflowDB := NewDB(
		"postgres",
		testConn,
		Logger(logger),
		Debug(),
	)

	workflowSvc := NewService(DB(workflowDB), Logger(logger), Debug())
	workflowHandler := ServiceHandler(ctx, workflowSvc)
	server := httptest.NewServer(workflowHandler)
	return server
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

// HTTP Test Code

var createWorkflowRequest = []byte(`{"name" :"http_test_workflow"}`)

func TestHTTPCreateWorkflow(t *testing.T) {
	req, err := client.NewRequest("workflows", "", jsonMedia, "POST")
	if err != nil {
		t.Fatal(err)
	}
	body := createWorkflowRequest
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
