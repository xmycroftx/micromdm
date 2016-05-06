package checkin

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/fullsailor/pkcs7"
	"github.com/groob/plist"
)

func decodeMDMCheckinRequest(r *http.Request) (interface{}, error) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(data))
	var request mdmCheckinRequest
	if err := plist.Unmarshal(data, &request); err != nil {
		return nil, err
	}
	return request, nil
}

// The enrollment request is PkCS7 signed.
// We'll ignore everything but the content for now
func decodeMDMEnrollmentRequest(r *http.Request) (interface{}, error) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	p7, err := pkcs7.Parse(data)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	// TODO: We should verify but not currently possible. Apple
	// does no provide a cert for the CA.
	var request depEnrollmentRequest
	if err := plist.Unmarshal(p7.Content, &request); err != nil {
		log.Println(err)
		return nil, err
	}
	return request, nil
}

// errorer is implemented by all concrete response types. It allows us to
// change the HTTP response code without needing to trigger an endpoint
// (transport-level) error. For more information, read the big comment in
// endpoint.go.
type errorer interface {
	error() error
}

// encodeResponse is the common method to encode all response types to the
// client. I chose to do it this way because I didn't know if something more
// specific was necessary. It's certainly possible to specialize on a
// per-response (per-method) basis.
func encodeResponse(w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(w, e.error())
		return nil
	}
	enc := plist.NewEncoder(w)
	enc.Indent("  ")
	return enc.Encode(response)
}

func enrollResponse(w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(w, e.error())
		return nil
	}
	resp := response.(depEnrollmentResponse)
	w.Write(resp.Profile)
	return nil
}

func encodeError(w http.ResponseWriter, err error) {
	w.WriteHeader(codeFrom(err))
	response := map[string]interface{}{
		"error": err.Error(),
	}
	enc := plist.NewEncoder(w)
	enc.Indent("  ")
	err = enc.Encode(response)
	if err != nil {
		log.Println(err)
	}
}

func codeFrom(err error) int {
	switch err {
	default:
		return http.StatusInternalServerError
	}
}
