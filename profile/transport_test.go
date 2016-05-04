package profile

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"testing/quick"
)

// Test sending POST requests to /mdm/profiles
// Then test sending a duplicate request
func TestHTTPCreateProfile(t *testing.T) {
	t.Log("generate profiles and send CreateProfile request to endpoint")
	assertion := func(pf Profile) bool {
		//encode generated body
		body, err := json.Marshal(pf)
		if err != nil {
			t.Fatal(err)
		}
		// create a request
		req, err := client.NewRequest("profiles", "", jsonMedia, "POST")
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
		}

		// now test duplicate
		req, err = client.NewRequest("profiles", "", jsonMedia, "POST")
		if err != nil {
			t.Fatal(err)
		}

		req.Body = &nopCloser{bytes.NewBuffer(body)}
		resp, err = client.Do(req, nil)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Error("Expected", http.StatusConflict, "got", resp.StatusCode)
		}
		return true
	}

	// quick check
	if err := quick.Check(assertion, nil); err != nil {
		t.Error(err)
	}
}

// Test sending GET requests to /mdm/profiles
// Should return HTTP code 200 and a list of Profiles
func TestHTTPListProfiles(t *testing.T) {
	req, err := client.NewRequest("profiles", "", jsonMedia, "GET")
	if err != nil {
		t.Fatalf("could not create GET request to /mdm/profiles: \n %q", err)
	}

	resp, err := client.Do(req, nil)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatal("expected", http.StatusOK, "got", resp.StatusCode)
	}

	var profiles []Profile
	err = json.NewDecoder(resp.Body).Decode(&profiles)
	if err != nil {
		t.Fatal(err)
	}

}
