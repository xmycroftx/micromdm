package connect

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/groob/plist"
)

func decodeMDMConnectRequest(r *http.Request) (interface{}, error) {
	body, _ := ioutil.ReadAll(r.Body)
	fmt.Println(string(body))
	reader := bytes.NewReader(body)

	var request mdmConnectRequest
	if err := plist.NewDecoder(reader).Decode(&request); err != nil {
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
	resp := response.(mdmConnectResponse)
	next := resp.payload
	// var decoded = make([]byte, base64.StdEncoding.DecodedLen(len(next)))
	// _, err := base64.StdEncoding.Decode(decoded, next)
	// if err != nil {
	// 	encodeError(w, err)
	// 	return nil
	// }
	if len(next) != 0 {
		w.Write(next)
	}
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
