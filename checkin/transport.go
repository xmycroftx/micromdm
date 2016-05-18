package checkin

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"

	"github.com/fullsailor/pkcs7"
	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/groob/plist"
)

// ServiceHandler returns an HTTP Handler for the checkin service
func ServiceHandler(ctx context.Context, svc Service, logger kitlog.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(encodeError),
	}

	checkinHandler := kithttp.NewServer(
		ctx,
		makeCheckinEndpoint(svc),
		decodeMDMCheckinRequest,
		encodeResponse,
		opts...,
	)
	depEnrollmentHandler := kithttp.NewServer(
		ctx,
		makeDEPEnrollmentEndpoint(svc),
		decodeMDMEnrollmentRequest,
		encodeDEPEnrollmentResponse,
		opts...,
	)
	r := mux.NewRouter()

	r.Handle("/mdm/checkin", checkinHandler).Methods("PUT")
	r.Handle("/mdm/checkin", depEnrollmentHandler).Methods("POST")
	return r
}

func decodeMDMCheckinRequest(_ context.Context, r *http.Request) (interface{}, error) {
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
func decodeMDMEnrollmentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	p7, err := pkcs7.Parse(data)
	if err != nil {
		return nil, err
	}
	// TODO: We should verify but not currently possible. Apple
	// does no provide a cert for the CA.
	var request depEnrollmentRequest
	if err := plist.Unmarshal(p7.Content, &request); err != nil {
		return nil, err
	}
	return request, nil
}

type errorer interface {
	error() error
}

type statuser interface {
	status() int
}

type listEncoder interface {
	encodeList(w http.ResponseWriter) error
}

func encodeDEPEnrollmentResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	resp := response.(depEnrollmentResponse)
	w.Write(resp.Profile)
	return nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	// w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// for success responses
	if e, ok := response.(statuser); ok {
		w.WriteHeader(e.status())
		if e.status() == http.StatusNoContent {
			return nil
		}
	}

	// check if this is a collection
	if e, ok := response.(listEncoder); ok {
		return e.encodeList(w)

	}
	return plist.NewEncoder(w).Encode(response)
}

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	// unwrap if the error is wrapped by kit http in it's own error type
	if httperr, ok := err.(kithttp.Error); ok {
		err = httperr.Err
	}
	switch err {
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	// w.Header().Set("Content-Type", "application/json; charset=utf-8")
	plist.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
