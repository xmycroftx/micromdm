package profile

import (
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"testing"
	"testing/quick"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
)

func TestNewDB(t *testing.T) {
	defer teardown()
	logger := log.NewLogfmtLogger(os.Stderr)
	_, err := NewDB("postgres", testConn, logger)
	if err != nil {
		t.Fatal(err)
	}
}

// TestAdd generates a bunch of new profiles
// and adds them to the database.
// The test checks that the profile stored is the same as the one returned.
func TestDatastoreAdd(t *testing.T) {
	ds := datastore(t)
	defer teardown()

	assertion := func(pf Profile) bool {
		newPrf, err := ds.Add(&pf)
		if err != nil {
			t.Fatal(err)
			return false
		}
		if newPrf.UUID == "" || newPrf.PayloadIdentifier != pf.PayloadIdentifier {
			return false
		}
		return true
	}

	if err := quick.Check(assertion, nil); err != nil {
		t.Error(err)
	}
}

func TestDatastoreProfiles(t *testing.T) {
	ds := datastore(t)
	defer teardown()

	// test with empty table
	testReturnNoArgs(t, ds, 0)

	// insert some profiles and retrieve them
	var profileCount = 5
	addRandomProfiles(profileCount, ds, t)
	profiles, err := ds.Profiles()
	if err != nil {
		t.Fatal(err)
	}

	returnedCount := len(profiles)
	if returnedCount != profileCount {
		t.Fatal("expected", profileCount, "got", returnedCount)
	}

	// build data for remaining tests
	var allIdentifiers []string
	var allUUIDs []string
	for _, p := range profiles {
		allIdentifiers = append(allIdentifiers, p.PayloadIdentifier)
		allUUIDs = append(allUUIDs, p.UUID)
	}

	// return everything
	testReturnNoArgs(t, ds, profileCount)
	// test nonexisting profile
	testReturnNone(t, ds)
	// test returning a unique id
	testReturnOneByIdentifier(t, ds, allIdentifiers[0])
	// test returning a unique profile by UUID
	testReturnOneByUUID(t, ds, allUUIDs[0])

}
func testReturnNoArgs(t *testing.T, ds Datastore, expectCount int) {
	profiles, err := ds.Profiles()
	if err != nil {
		t.Fatal(err)
	}

	returnedCount := len(profiles)
	if returnedCount != expectCount {
		t.Fatal("expected", expectCount, "got", returnedCount)
	}
}

func testReturnNone(t *testing.T, ds Datastore) {
	identifier := "com.example.does.not.exist"
	profiles, err := ds.Profiles(Identifier{identifier})
	if err != nil {
		t.Fatal(err)
	}

	returnedCount := len(profiles)
	if returnedCount != 0 {
		t.Fatal("expected", 0, "got", returnedCount)
	}
}

func testReturnOneByIdentifier(t *testing.T, ds Datastore, id string) {
	profiles, err := ds.Profiles(Identifier{id})
	if err != nil {
		t.Fatal(err)
	}

	returnedCount := len(profiles)
	if returnedCount != 1 {
		t.Fatal("expected", 1, "got", returnedCount)
	}

	returnedIdentifier := profiles[0].PayloadIdentifier
	if returnedIdentifier != id {
		t.Fatal("expected", id, "got", returnedIdentifier)
	}
}

func testReturnOneByUUID(t *testing.T, ds Datastore, uuid string) {
	profiles, err := ds.Profiles(UUID{uuid})
	if err != nil {
		t.Fatal(err)
	}

	returnedCount := len(profiles)
	if returnedCount != 1 {
		t.Fatal("expected", 1, "got", returnedCount)
	}

	returnedUUID := profiles[0].UUID
	if returnedUUID != uuid {
		t.Fatal("expected", uuid, "got", returnedUUID)
	}
}

func addRandomProfiles(num int, ds Datastore, t *testing.T) {
	for i := 1; i <= num; i++ {
		pf := randomProfile()
		_, err := ds.Add(&pf)
		if err != nil {
			t.Fatal(err)
		}
	}
}

// helpers
func randomProfile() Profile {
	vrf, ok := quick.Value(reflect.TypeOf(Profile{}), rand.New(rand.NewSource(1)))
	if !ok {
		panic("randomProfile: no value")
	}

	if f, ok := vrf.Interface().(Profile); ok {
		return f
	}
	return Profile{}
}

// implement quick.Generator for Profile
func (pf Profile) Generate(rand *rand.Rand, size int) reflect.Value {
	a := RandomString(16)
	b := RandomString(16)
	c := RandomString(16)
	randomIdnetifier := fmt.Sprintf("%v.%v.%v", a, b, c)
	randomProfile := Profile{
		PayloadIdentifier: randomIdnetifier,
		ProfileData:       `PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiPz4KPCFET0NUWVBFIHBsaXN0IFBVQkxJQyAiLS8vQXBwbGUvL0RURCBQTElTVCAxLjAvL0VOIiAiaHR0cDovL3d3dy5hcHBsZS5jb20vRFREcy9Qcm9wZXJ0eUxpc3QtMS4wLmR0ZCI+CjxwbGlzdCB2ZXJzaW9uPSIxLjAiPgo8ZGljdD4KICAgIDxrZXk+UGF5bG9hZENvbnRlbnQ8L2tleT4KICAgIDxhcnJheT4KICAgICAgICA8ZGljdD4KICAgICAgICAgICAgPGtleT5QYXlsb2FkQ29udGVudDwva2V5PgogICAgICAgICAgICA8ZGljdD4KICAgICAgICAgICAgICAgIDxrZXk+Y29tLmFwcGxlLlNldHVwQXNzaXN0YW50PC9rZXk+CiAgICAgICAgICAgICAgICA8ZGljdD4KICAgICAgICAgICAgICAgICAgICA8a2V5PlNldC1PbmNlPC9rZXk+CiAgICAgICAgICAgICAgICAgICAgPGFycmF5PgogICAgICAgICAgICAgICAgICAgICAgICA8ZGljdD4KICAgICAgICAgICAgICAgICAgICAgICAgICAgIDxrZXk+bWN4X2RhdGFfdGltZXN0YW1wPC9rZXk+CiAgICAgICAgICAgICAgICAgICAgICAgICAgICA8ZGF0ZT4yMDE0LTEwLTI5VDE3OjIwOjEwWjwvZGF0ZT4KICAgICAgICAgICAgICAgICAgICAgICAgICAgIDxrZXk+bWN4X3ByZWZlcmVuY2Vfc2V0dGluZ3M8L2tleT4KICAgICAgICAgICAgICAgICAgICAgICAgICAgIDxkaWN0PgogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIDxrZXk+RGlkU2VlQ2xvdWRTZXR1cDwva2V5PgogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIDx0cnVlLz4KICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICA8a2V5Pkdlc3R1cmVNb3ZpZVNlZW48L2tleT4KICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICA8c3RyaW5nPm5vbmU8L3N0cmluZz4KICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICA8a2V5Pkxhc3RTZWVuQ2xvdWRQcm9kdWN0VmVyc2lvbjwva2V5PgogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIDxzdHJpbmc+MTAuMTEuMjwvc3RyaW5nPgogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIDxrZXk+TGFzdFNlZW5CdWRkeUJ1aWxkVmVyc2lvbjwva2V5PgogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIDxzdHJpbmc+MTVDNTA8L3N0cmluZz4KICAgICAgICAgICAgICAgICAgICAgICAgICAgIDwvZGljdD4KICAgICAgICAgICAgICAgICAgICAgICAgPC9kaWN0PgogICAgICAgICAgICAgICAgICAgIDwvYXJyYXk+CiAgICAgICAgICAgICAgICA8L2RpY3Q+CiAgICAgICAgICAgIDwvZGljdD4KICAgICAgICAgICAgPGtleT5QYXlsb2FkRW5hYmxlZDwva2V5PgogICAgICAgICAgICA8dHJ1ZS8+CiAgICAgICAgICAgIDxrZXk+UGF5bG9hZElkZW50aWZpZXI8L2tleT4KICAgICAgICAgICAgPHN0cmluZz5lZHUucHJhdHQuc3VwcHJlc3NfaWNsb3VkX2Fzc3Q8L3N0cmluZz4KICAgICAgICAgICAgPGtleT5QYXlsb2FkVHlwZTwva2V5PgogICAgICAgICAgICA8c3RyaW5nPmNvbS5hcHBsZS5NYW5hZ2VkQ2xpZW50LnByZWZlcmVuY2VzPC9zdHJpbmc+CiAgICAgICAgICAgIDxrZXk+UGF5bG9hZFVVSUQ8L2tleT4KICAgICAgICAgICAgPHN0cmluZz41ZTEwZjM0OC05MjNiLTQzOGEtOWI4Ny1mYTk3OGU4NmUxMWE8L3N0cmluZz4KICAgICAgICAgICAgPGtleT5QYXlsb2FkVmVyc2lvbjwva2V5PgogICAgICAgICAgICA8aW50ZWdlcj4xPC9pbnRlZ2VyPgogICAgICAgIDwvZGljdD4KICAgIDwvYXJyYXk+CiAgICA8a2V5PlBheWxvYWREZXNjcmlwdGlvbjwva2V5PgogICAgPHN0cmluZz5Db25maWd1cmVzIGNvbS5hcHBsZS5TZXR1cEFzc2lzdGFudDwvc3RyaW5nPgogICAgPGtleT5QYXlsb2FkRGlzcGxheU5hbWU8L2tleT4KICAgIDxzdHJpbmc+aUNsb3VkIFNldHVwQXNzaXN0YW50IENvbmZpZ3VyYXRpb248L3N0cmluZz4KICAgIDxrZXk+UGF5bG9hZElkZW50aWZpZXI8L2tleT4KICAgIDxzdHJpbmc+Y29tLmdpdGh1Yi5ncmVnbmVhZ2xlLnN1cHByZXNzX2ljbG91ZF9hc3N0PC9zdHJpbmc+CiAgICA8a2V5PlBheWxvYWRPcmdhbml6YXRpb248L2tleT4KICAgIDxzdHJpbmc+PC9zdHJpbmc+CiAgICA8a2V5PlBheWxvYWRSZW1vdmFsRGlzYWxsb3dlZDwva2V5PgogICAgPGZhbHNlLz4KICAgIDxrZXk+UGF5bG9hZFNjb3BlPC9rZXk+CiAgICA8c3RyaW5nPlN5c3RlbTwvc3RyaW5nPgogICAgPGtleT5QYXlsb2FkVHlwZTwva2V5PgogICAgPHN0cmluZz5Db25maWd1cmF0aW9uPC9zdHJpbmc+CiAgICA8a2V5PlBheWxvYWRVVUlEPC9rZXk+CiAgICA8c3RyaW5nPmU4MWY1MWMyLTExODAtNGRlMC05NGNkLTMxNTNhYTQxMzg3Njwvc3RyaW5nPgogICAgPGtleT5QYXlsb2FkVmVyc2lvbjwva2V5PgogICAgPGludGVnZXI+MTwvaW50ZWdlcj4KPC9kaWN0Pgo8L3BsaXN0Pgo=`,
	}
	return reflect.ValueOf(randomProfile)
}

func RandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

var (
	testConn = "user=micromdm password=micromdm dbname=micromdm sslmode=disable"
)

func datastore(t *testing.T) Datastore {
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
	DROP TABLE IF EXISTS profiles;
	`
	db.MustExec(drop)
	defer db.Close()
}
