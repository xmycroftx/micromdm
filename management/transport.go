package management

import (
	"encoding/json"
	"errors"
	"net/http"

	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
)

// ServiceHandler returns an HTTP Handler for the management service
func ServiceHandler(ctx context.Context, svc Service, logger kitlog.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(encodeError),
	}

	fetchDEPHandler := kithttp.NewServer(
		ctx,
		makeFetchDevicesEndpoint(svc),
		decodeFetchDEPDevicesRequest,
		encodeResponse,
		opts...,
	)

	r := mux.NewRouter()

	r.Handle("/management/v1/devices/fetch", fetchDEPHandler).Methods("POST")

	return r
}

var errBadRoute = errors.New("bad route")

func decodeFetchDEPDevicesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return fetchDEPDevicesRequest{}, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

type errorer interface {
	error() error
}

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	switch err {
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
