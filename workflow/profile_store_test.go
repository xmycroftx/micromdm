package workflow

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

func TestDeleteProfile(t *testing.T) {
	ds := datastore(t)
	defer teardown()
	testProfiles := addTestProfiles(t, ds, 5)

	for _, p := range testProfiles {
		err := ds.DeleteProfile(&p)
		if err != nil {
			t.Fatal(err)
		}

	}

	empty := Profile{}
	err := ds.DeleteProfile(&empty)
	if err != nil {
		t.Error(err)
	}
	badUUIDProfile := Profile{
		UUID:              "bad.uuid",
		PayloadIdentifier: "with.bad.uuid",
	}
	err = ds.DeleteProfile(&badUUIDProfile)
	if err == nil {
		t.Fatal("expected an error but got nil")
	}

}

func TestRetrieveProfiles(t *testing.T) {
	ds := datastore(t)
	defer teardown()
	testProfiles := addTestProfiles(t, ds, 5)
	// retrieve all
	profiles, err := ds.Profiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 5 {
		t.Error("expected", 5, "got", len(profiles))
	}

	for _, p := range testProfiles {
		byUUID, err := ds.Profiles(ProfileUUID{p.UUID})
		if err != nil {
			t.Fatal(err)
		}
		if len(byUUID) != 1 {
			t.Log("filtering by UUID should only return 1 result")
			t.Fatal("expected", 1, "got", len(byUUID))
		}

		uuid := byUUID[0].UUID
		if p.UUID != uuid {
			t.Log("result should have the same UUID as the one in the query")
			t.Fatal("expected", p.UUID, "got", uuid)

		}

		byPayloadIdentifier, err := ds.Profiles(PayloadIdentifier{p.PayloadIdentifier})
		if err != nil {
			t.Fatal(err)
		}
		if len(byPayloadIdentifier) != 1 {
			t.Log("filtering by PayloadIdentifier should only return 1 result")
			t.Fatal("expected", 1, "got", len(byPayloadIdentifier))
		}
	}

	badUUIDQuery := ProfileUUID{"bad_uuid"}
	_, err = ds.Profiles(badUUIDQuery)
	if err == nil {
		t.Fatal("expected an error but got nil")

	}
}

// Generates new Profile types and stores them in the datastore
func TestCreateProfile(t *testing.T) {
	ds := datastore(t)
	defer teardown()

	assertion := func(pf Profile) bool {
		newPrf, err := ds.CreateProfile(&pf)
		if err != nil {
			t.Fatal(err)
			return false
		}
		if newPrf.UUID == "" || newPrf.PayloadIdentifier != pf.PayloadIdentifier {
			return false
		}

		// now try duplicates

		_, err = ds.CreateProfile(&pf)
		if err != ErrExists || err == nil {
			t.Log("ds should not create duplicate resources")
			t.Fatal("expected", ErrExists, "got", err)
			return false
		}
		return true
	}

	if err := quick.Check(assertion, nil); err != nil {
		t.Error(err)
	}

	// empty profile test
	empty := Profile{}
	_, err := ds.CreateProfile(&empty)
	if err == nil {
		t.Fatal("expected an error but got nil")
	}
}

// add some profiles to the datastore for quick testing
func addTestProfiles(t *testing.T, ds Datastore, numProfiles int) []Profile {
	var profiles []Profile
	for i := 0; i < numProfiles; i++ {
		input := randomProfile()
		newProfile, err := ds.CreateProfile(&input)
		if err != nil {
			t.Fatal(err)
		}
		profiles = append(profiles, *newProfile)
	}
	return profiles
}

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

// Generate a random profile
func (pf Profile) Generate(rand *rand.Rand, size int) reflect.Value {
	a := randomString(3)
	b := randomString(10)
	c := randomString(5)
	data, err := ioutil.ReadFile("testdata/sample_profile")
	if err != nil {
		panic(err)
	}
	randomIdnetifier := fmt.Sprintf("%v.%v.%v", a, b, c)
	randomProfile := Profile{
		PayloadIdentifier: randomIdnetifier,
		ProfileData:       string(data),
	}
	return reflect.ValueOf(randomProfile)

}
