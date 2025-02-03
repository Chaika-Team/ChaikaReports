package http

import (
	"ChaikaReports/internal/handler/http/decoder"
	"ChaikaReports/internal/handler/http/encoder"
	"ChaikaReports/internal/service"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// NewHTTPHandler initializes and returns a new HTTP handler with all routes defined
func NewHTTPHandler(svc service.SalesService) http.Handler {
	r := mux.NewRouter()

	// Serve Swagger UI at /docs/
	r.PathPrefix("/docs/").Handler(httpSwagger.WrapHandler)

	apiV1 := r.PathPrefix("/api/v1").Subrouter()

	// Register the insert sales route
	apiV1.Handle("/sales", httptransport.NewServer(
		MakeInsertSalesEndpoint(svc),
		decoder.DecodeInsertSalesRequest,
		encoder.EncodeResponse,
		httptransport.ServerErrorEncoder(encoder.EncodeError),
	)).Methods("POST")

	// Add more routes as needed

	return r
}
