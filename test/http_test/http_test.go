package http_test

import (
	httphandler "ChaikaReports/internal/handler/http"
	"ChaikaReports/internal/handler/http/schemas"
	"ChaikaReports/internal/models"
	"ChaikaReports/internal/service"
	"bytes"
	"context"
	"encoding/json"
	"github.com/go-kit/log"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSalesRepository is a mock implementation of the repository.SalesRepository interface
type MockSalesRepository struct {
	mock.Mock
}

func (m *MockSalesRepository) InsertData(ctx context.Context, carriageReport *models.Carriage) error {
	args := m.Called(ctx, carriageReport)
	return args.Error(0)
}

func (m *MockSalesRepository) GetEmployeeCartsInTrip(ctx context.Context, tripID *models.TripID, employeeID *string) ([]models.Cart, error) {
	args := m.Called(tripID, employeeID)
	if args.Get(0) != nil {
		return args.Get(0).([]models.Cart), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSalesRepository) GetEmployeeIDsByTrip(tripID *models.TripID) ([]string, error) {
	args := m.Called(tripID)
	if args.Get(0) != nil {
		return args.Get(0).([]string), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSalesRepository) UpdateItemQuantity(tripID *models.TripID, cartID *models.CartID, productID *int, newQuantity *int16) error {
	args := m.Called(tripID, cartID, productID, newQuantity)
	return args.Error(0)
}

func (m *MockSalesRepository) DeleteItemFromCart(tripID *models.TripID, cartID *models.CartID, productID *int) error {
	args := m.Called(tripID, cartID, productID)
	return args.Error(0)
}

func TestInsertSalesEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		rawJSON        string
		mockSetup      func(*MockSalesRepository)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Successful Insert",
			rawJSON: `{
   		  "trip_id": {
   		    "route_id": "route_test",
   		    "start_time": "2023-01-15T10:00:01Z"
   		  },
   		  "end_time": "2023-01-15T11:00:01Z",
   		  "carriage_id": 10,
   		  "carts": [
   		    {
   		      "cart_id": {
   		        "employee_id": "67890",
   		        "operation_time": "2023-01-15T12:30:00Z"
   		      },
   		      "operation_type": 1,
   		      "items": [
   		        {"product_id": 1, "quantity": 10, "price": 100.0},
   		        {"product_id": 2, "quantity": 5, "price": 200.0},
   		        {"product_id": 3, "quantity": 2, "price": 250.0}
   		      ]
   		    }
   		  ]
   		}`,
			mockSetup: func(m *MockSalesRepository) {
				m.On("InsertData", mock.Anything, mock.AnythingOfType("*models.Carriage")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: schemas.InsertSalesResponse{
				Message: "Data inserted successfully",
			},
		},
		{
			name: "Malformed JSON",
			rawJSON: `{
   		  "trip_id": {
   		    "route_id": "route_test"
   		    "start_time": "2023-01-15T10:00:01Z"
   		  },
   		  "end_time": "2023-01-15T11:00:01Z",
   		  "carriage_id": 10,
   		  "carts": []
   		}`, // Missing comma between route_id and start_time
			mockSetup:      func(m *MockSalesRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: schemas.ErrorResponse{
				Error: "invalid request body",
			},
		},
		{
			name: "Incorrect quantity",
			rawJSON: `{
   		  "trip_id": {
   		    "route_id": "route_test",
   		    "start_time": "2023-01-15T10:00:01Z"
   		  },
   		  "end_time": "2023-01-15T11:00:01Z",
   		  "carriage_id": 10,
   		  "carts": [
   		    {
   		      "cart_id": {
   		        "employee_id": "67890",
   		        "operation_time": "2023-01-15T12:30:00Z"
   		      },
   		      "operation_type": 1,
   		      "items": [
   		        {"product_id": 1, "quantity": 0, "price": 100.0},
   		        {"product_id": 2, "quantity": 5, "price": 200.0},
   		        {"product_id": 3, "quantity": 2, "price": 250.0}
   		      ]
   		    }
   		  ]
   		}`, // Missing comma between route_id and start_time
			mockSetup:      func(m *MockSalesRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: schemas.ErrorResponse{
				Error: "invalid item quantity",
			},
		},
		{
			name: "Missing RouteID",
			rawJSON: `{
   		  "trip_id": {
   		    "start_time": "2023-01-15T10:00:01Z"
   		  },
   		  "end_time": "2023-01-15T11:00:01Z",
   		  "carriage_id": 10,
   		  "carts": []
   		}`, // Missing trip_id
			mockSetup:      func(m *MockSalesRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: schemas.ErrorResponse{
				Error: "validation failed: Key: 'InsertSalesRequest.TripID.RouteID' Error:Field validation for 'RouteID' failed on the 'required' tag",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize mock repository
			mockRepo := &MockSalesRepository{}
			tt.mockSetup(mockRepo)

			// Initialize service with mock repository
			svc := service.NewSalesService(mockRepo)

			// Initialize HTTP handler with the service
			handler := httphandler.NewHTTPHandler(svc, log.NewNopLogger())

			// Create HTTP request
			req, err := http.NewRequest("POST", "/api/v1/sales", bytes.NewBufferString(tt.rawJSON))
			assert.NoError(t, err, "Failed to create new request")
			req.Header.Set("Content-Type", "application/json")

			// Create ResponseRecorder
			rr := httptest.NewRecorder()

			// Serve the request
			handler.ServeHTTP(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code, "Unexpected status code")

			// Decode and check response body
			body, err := io.ReadAll(rr.Body)
			assert.NoError(t, err, "Failed to read response body")

			expectedBodyJSON, _ := json.Marshal(tt.expectedBody)
			assert.JSONEq(t, string(expectedBodyJSON), string(body), "Response body does not match")

			// Assert that InsertData was called if expected
			if tt.expectedStatus == http.StatusOK {
				mockRepo.AssertCalled(t, "InsertData", mock.Anything, mock.AnythingOfType("*models.Carriage"))
			} else {
				mockRepo.AssertNotCalled(t, "InsertData", mock.Anything, mock.Anything)
			}
		})
	}
}
