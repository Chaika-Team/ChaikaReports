package encoder

import (
	"ChaikaReports/internal/handler/http/schemas"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kit/log"
	"net/http"
)

// EncodeResponse encodes the domain response into an HTTP response
func EncodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	if res, ok := response.(schemas.InsertSalesResponse); ok {
		w.WriteHeader(http.StatusOK)
		return json.NewEncoder(w).Encode(res)
	}
	// Handle other response types if necessary
	return fmt.Errorf("unknown response type: %T", response)
}

// EncodeError encodes errors into an HTTP error response
func EncodeError(logger log.Logger) func(_ context.Context, err error, w http.ResponseWriter) {
	return func(_ context.Context, err error, w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		var code int
		var msg string

		switch {
		case isValidationError(err):
			code = http.StatusBadRequest
			msg = err.Error()
		default:
			code = http.StatusInternalServerError
			// Log the actual error but return a generic message
			_ = logger.Log("error", fmt.Sprintf("Internal server error: %v", err))
			msg = http.StatusText(http.StatusInternalServerError)
		}

		w.WriteHeader(code)
		if err := json.NewEncoder(w).Encode(schemas.ErrorResponse{Error: msg}); err != nil {
			_ = logger.Log("error", fmt.Sprintf("Failed to encode error response: %v", err))
		}
	}
}

// Helper function to determine if the error is a validation error
func isValidationError(err error) bool {
	// Implement logic to determine if err is a validation error
	// For simplicity, assume all errors returned from decoder are validation errors
	return true
}
