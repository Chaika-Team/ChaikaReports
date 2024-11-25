package http

import (
	"ChaikaReports/internal/handler/http/decoder"
	"ChaikaReports/internal/handler/http/encoder"
	"ChaikaReports/internal/service"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// NewHTTPHandler initializes and returns a new HTTP handler with all routes defined
func NewHTTPHandler(svc service.SalesService) http.Handler {
	r := mux.NewRouter()

	// Register the insert sales route
	r.Handle("/api/v1/sales", httptransport.NewServer(
		MakeInsertSalesEndpoint(svc),
		decoder.DecodeInsertSalesRequest,
		encoder.EncodeResponse,
		httptransport.ServerErrorEncoder(encoder.EncodeError),
	)).Methods("POST")

	// Add more routes as needed

	return r
}
