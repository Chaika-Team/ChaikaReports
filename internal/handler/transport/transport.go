package transport

import (
	"ChaikaReports/internal/handler/endpoints"
	"context"
	"encoding/json"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
)

func NewHTTPHandler(endpoints endpoints.Endpoints) http.Handler {
	m := http.NewServeMux()

	m.Handle("/insert", httptransport.NewServer(
		endpoints.InsertDataEndpoint,
		decodeInsertDataRequest,
		encodeResponse,
	))

	m.Handle("/actions/conductor", httptransport.NewServer(
		endpoints.GetActionsByConductorEndpoint,
		decodeGetActionsByConductorRequest,
		encodeResponse,
	))

	m.Handle("/conductors", httptransport.NewServer(
		endpoints.GetConductorsByTripIDEndpoint,
		decodeGetConductorsByTripIDRequest,
		encodeResponse,
	))

	m.Handle("/actions/update", httptransport.NewServer(
		endpoints.UpdateActionCountEndpoint,
		decodeUpdateActionCountRequest,
		encodeResponse,
	))

	m.Handle("/actions/delete", httptransport.NewServer(
		endpoints.DeleteProductFromActionEndpoint,
		decodeDeleteProductFromActionRequest,
		encodeResponse,
	))

	return m
}

// Decoders for each request

func decodeInsertDataRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoints.InsertDataRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func decodeGetActionsByConductorRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoints.GetActionsByConductorRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func decodeGetConductorsByTripIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoints.GetConductorsByTripIDRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func decodeUpdateActionCountRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoints.UpdateActionCountRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func decodeDeleteProductFromActionRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoints.DeleteProductFromActionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

// Encode response for all requests
func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}
