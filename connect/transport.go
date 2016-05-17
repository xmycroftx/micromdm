package connect

import (
	"net/http"

	"golang.org/x/net/context"

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

	connectHandler := kithttp.NewServer(
		ctx,
		makeConnectEndpoint(svc),
		decodeMDMConnectRequest,
		encodeResponse,
		opts...,
	)
	r := mux.NewRouter()

	r.Handle("/mdm/connect", connectHandler).Methods("PUT")
	return r
}

func decodeMDMConnectRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request mdmConnectRequest
	if err := plist.NewDecoder(r.Body).Decode(&request); err != nil {
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

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
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

	resp := response.(mdmConnectResponse)
	next := resp.payload
	if len(next) != 0 {
		w.Write(next)
	}
	return nil
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
	plist.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
