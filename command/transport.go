package command

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/gorilla/mux"
	"github.com/micromdm/mdm"
	"golang.org/x/net/context"
)

var (
	// ErrEmptyRequest is returned if the request body is empty
	ErrEmptyRequest = errors.New("request must contain UDID of the device")
	errBadRouting   = errors.New("inconsistent mapping between route and handler (programmer error)")
)

// NewCommandRequest represents an HTTP Request for a new MDM Command
type NewCommandRequest struct {
	*mdm.CommandRequest
}

func decodeNewCommandRequest(r *http.Request) (interface{}, error) {
	var request NewCommandRequest
	err := json.NewDecoder(r.Body).Decode(&request.CommandRequest)
	return request, err
}

// NewCommandResponse is a command reponse
type NewCommandResponse struct {
	*mdm.Payload
	Err error `json:"error,omitempty"`
}

func (r NewCommandResponse) error() error { return r.Err }

// errorer is implemented by all concrete response types. It allows us to
// change the HTTP response code without needing to trigger an endpoint
// (transport-level) error. For more information, read the big comment in
// endpoint.go.
type errorer interface {
	error() error
}

// NextCommandRequest is a request to return the next command in a device queue
type NextCommandRequest struct {
	UDID string
}

func decodeNextCommandRequest(r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	udid, ok := vars["udid"]
	if !ok {
		return nil, errBadRouting
	}
	var request NextCommandRequest
	request.UDID = udid
	return request, nil
}

// EncodeNextCommandRequest encodes a request for the NextCommand endpoint
func EncodeNextCommandRequest(r *http.Request, request interface{}) error {
	req := request.(NextCommandRequest)
	path := r.URL.Path
	r.URL.Path = fmt.Sprintf("%v/%v/next", path, req.UDID)
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)
	return nil
}

// NextCommandResponse is a response for the next command
type NextCommandResponse struct {
	Payload []byte `json:"command_payload"`
	Total   int    `json:"total_payloads"`
	Err     error  `json:"error,omitempty"`
}

// DecodeNextCommandResponse decodes the response from the provided HTTP response,
// simply by JSON decoding from the response body. It's designed to be used in
// transport/http.Client.
// first decode into map[string]interface{} and check for error in the response
func DecodeNextCommandResponse(resp *http.Response) (interface{}, error) {
	var r map[string]interface{}
	var response NextCommandResponse
	err := json.NewDecoder(resp.Body).Decode(&r)
	if rs, ok := r["error"]; ok {
		response.Err = errors.New(rs.(string))
	}
	if rs, ok := r["total_payloads"]; ok {
		response.Total = int(rs.(float64))
	}
	if rs, ok := r["command_payload"]; ok {
		response.Payload = []byte(rs.(string))
	}
	return response, err
}

func (r NextCommandResponse) error() error { return r.Err }

// DeleteCommandRequest is a request to delete a command
type DeleteCommandRequest struct {
	// device UDID
	UDID string
	// command UUID
	UUID string
}

// EncodeDeleteCommandRequest encodes a request for the NextCommand endpoint
func EncodeDeleteCommandRequest(r *http.Request, request interface{}) error {
	req := request.(DeleteCommandRequest)
	path := r.URL.Path
	r.URL.Path = fmt.Sprintf("%v/%v/%v", path, req.UDID, req.UUID)
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)
	return nil
}

func decodeDeleteCommandRequest(r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	udid, ok := vars["udid"]
	if !ok {
		return nil, errBadRouting
	}
	uuid, ok := vars["uuid"]
	if !ok {
		return nil, errBadRouting
	}
	var request DeleteCommandRequest
	request.UDID = udid
	request.UUID = uuid
	return request, nil
}

// DeleteCommandResponse is a response for a delete request
type DeleteCommandResponse struct {
	Total int   `json:"remaining_payloads"`
	Err   error `json:"error,omitempty"`
}

// DecodeDeleteCommandResponse decodes the response from the provided HTTP response,
// simply by JSON decoding from the response body. It's designed to be used in
// transport/http.Client.
// first decode into map[string]interface{} and check for error in the response
func DecodeDeleteCommandResponse(resp *http.Response) (interface{}, error) {
	var r map[string]interface{}
	var response DeleteCommandResponse
	err := json.NewDecoder(resp.Body).Decode(&r)
	if rs, ok := r["error"]; ok {
		response.Err = errors.New(rs.(string))
	}
	if rs, ok := r["remaining_payloads"]; ok {
		response.Total = int(rs.(float64))
	}
	return response, err
}

func makeNewCommandEndpoint(svc MDMCommandService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NewCommandRequest)
		if req.UDID == "" || req.RequestType == "" {
			return NewCommandResponse{Err: ErrEmptyRequest}, nil
		}
		payload, err := svc.NewCommand(req.CommandRequest)
		if err != nil {
			return NewCommandResponse{Err: err}, nil
		}
		return NewCommandResponse{Payload: payload}, nil
	}
}

func makeNextCommandEndpoint(svc MDMCommandService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NextCommandRequest)
		if req.UDID == "" {
			return NextCommandResponse{Err: ErrEmptyRequest}, nil
		}
		payload, total, err := svc.NextCommand(req.UDID)
		if err != nil {
			return NextCommandResponse{Err: err}, nil
		}
		return NextCommandResponse{Payload: payload, Total: total}, nil
	}
}

func makeDeleteCommandEndpoint(svc MDMCommandService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DeleteCommandRequest)
		if req.UDID == "" {
			return DeleteCommandResponse{Err: ErrEmptyRequest}, nil
		}
		if req.UUID == "" {
			return DeleteCommandResponse{Err: ErrEmptyRequest}, nil
		}
		total, err := svc.DeleteCommand(req.UDID, req.UUID)
		if err != nil {
			return DeleteCommandResponse{Err: err}, nil
		}
		return DeleteCommandResponse{Total: total}, nil
	}
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
