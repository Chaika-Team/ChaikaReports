package http

import (
	"ChaikaReports/internal/models"
	"context"
	"encoding/json"
	"github.com/go-kit/kit/endpoint"
	httpGoKit "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
)

func NewHTTPServer(logger log.Logger, endpoints Endpoints) http.Handler {
	r := mux.NewRouter()
	r.Use(commonMiddleware(logger))

	// Swagger UI
	r.PathPrefix("/docs/").Handler(httpSwagger.WrapHandler)

	// Apply logging middleware to all endpoints
	wrapEndpointsWithLogging(logger, &endpoints)

	// Register routes
	registerRoutes(logger, r, endpoints)

	return r
}

func commonMiddleware(logger log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = level.Info(logger).Log("msg", "received request", "method", r.Method, "url", r.URL.String())
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, r)
			_ = level.Info(logger).Log("msg", "handled request", "method", r.Method, "url", r.URL.String())
		})
	}
}

func wrapEndpointsWithLogging(logger log.Logger, endpoints *Endpoints) {
	loggingMiddleware := makeLoggingMiddleware(logger)
	endpoints.InsertData = loggingMiddleware(endpoints.InsertData)
	// Add more endpoints here if needed
}

func makeLoggingMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			_ = level.Info(logger).Log("msg", "calling endpoint", "request", request)
			response, err = next(ctx, request)
			_ = level.Info(logger).Log("msg", "called endpoint", "response", response, "err", err)
			return
		}
	}
}

func registerRoutes(logger log.Logger, r *mux.Router, endpoints Endpoints) {
	api := r.PathPrefix("/api/v1").Subrouter()

	// Register the InsertData endpoint
	api.Methods("POST").Path("/sales").Handler(httpGoKit.NewServer(
		endpoints.InsertData,
		decodeInsertDataRequest,
		encodeResponse(logger),
		httpGoKit.ServerErrorEncoder(encodeErrorResponse(logger)),
	))
}

func decodeInsertDataRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req models.Carriage
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func encodeResponse(logger log.Logger) httpGoKit.EncodeResponseFunc {
	return func(ctx context.Context, w http.ResponseWriter, response interface{}) error {
		w.WriteHeader(http.StatusOK)
		return json.NewEncoder(w).Encode(response)
	}
}

func encodeErrorResponse(logger log.Logger) httpGoKit.ErrorEncoder {
	return func(ctx context.Context, err error, w http.ResponseWriter) {
		w.WriteHeader(determineHTTPError(err))
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
	}
}

func determineHTTPError(err error) int {
	return http.StatusInternalServerError
}
