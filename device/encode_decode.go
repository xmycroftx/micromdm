package device

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var errBadRoute = errors.New("bad route")

func decodeAssignWorkflowRequest(r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	udid, ok := vars["udid"]
	if !ok {
		return nil, errBadRoute
	}
	request := assignWorkflowRequest{DeviceUDID: udid}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
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
	jsn, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return err
	}
	w.Write(jsn)
	return nil
}

func encodeError(w http.ResponseWriter, err error) {
	w.WriteHeader(codeFrom(err))
	response := map[string]interface{}{
		"error": err.Error(),
	}
	jsn, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Println(err)
		return
	}
	w.Write(jsn)
}

func codeFrom(err error) int {
	switch err {
	default:
		return http.StatusInternalServerError
	}
}
