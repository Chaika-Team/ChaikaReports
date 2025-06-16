package http_test

import (
	httphandler "ChaikaReports/internal/handler/http"
	"ChaikaReports/internal/handler/http/schemas"
	"ChaikaReports/internal/models"
	"ChaikaReports/internal/service"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-kit/log"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSalesRepository is a mock implementation of the repository.SalesRepository interface
type MockSalesRepository struct {
	mock.Mock
}

func (m *MockSalesRepository) GetUnsyncedTrips(ctx context.Context) ([]models.TripID, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockSalesRepository) DeleteSyncedTrip(ctx context.Context, routeID string, startTime time.Time) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockSalesRepository) InsertData(ctx context.Context, carriageReport *models.CarriageReport) error {
	args := m.Called(ctx, carriageReport)
	return args.Error(0)
}

func (m *MockSalesRepository) GetTrip(ctx context.Context, tripID *models.TripID) (models.Trip, error) {
	args := m.Called(ctx, tripID)
	// if the first argument isn't nil and can be asserted to models.Trip, return it
	if trip, ok := args.Get(0).(models.Trip); ok {
		return trip, args.Error(1)
	}
	// otherwise return the zero‐value of models.Trip
	return models.Trip{}, args.Error(1)
}

func (m *MockSalesRepository) GetEmployeeCartsInTrip(ctx context.Context, tripID *models.TripID, employeeID *string) ([]models.Cart, error) {
	args := m.Called(ctx, tripID, employeeID)
	if args.Get(0) != nil {
		return args.Get(0).([]models.Cart), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSalesRepository) GetEmployeeIDsByTrip(ctx context.Context, tripID *models.TripID) ([]string, error) {
	args := m.Called(ctx, tripID)
	if args.Get(0) != nil {
		return args.Get(0).([]string), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSalesRepository) GetEmployeeTrips(ctx context.Context, employeeID string, year string) ([]models.EmployeeTrip, error) {
	args := m.Called(ctx, employeeID, year)
	if args.Get(0) != nil {
		return args.Get(0).([]models.EmployeeTrip), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSalesRepository) UpdateItemQuantity(ctx context.Context, tripID *models.TripID, cartID *models.CartID, productID *int, newQuantity *int16) error {
	args := m.Called(ctx, tripID, cartID, productID, newQuantity)
	return args.Error(0)
}

func (m *MockSalesRepository) DeleteItemFromCart(ctx context.Context, tripID *models.TripID, cartID *models.CartID, productID *int) error {
	args := m.Called(ctx, tripID, cartID, productID)
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
   		        {"product_id": 1, "quantity": 10, "price": 100},
   		        {"product_id": 2, "quantity": 5, "price": 200},
   		        {"product_id": 3, "quantity": 2, "price": 250}
   		      ]
   		    }
   		  ]
   		}`,
			mockSetup: func(m *MockSalesRepository) {
				m.On("InsertData", mock.Anything, mock.AnythingOfType("*models.CarriageReport")).Return(nil)
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
   		        {"product_id": 1, "quantity": 0, "price": 100},
   		        {"product_id": 2, "quantity": 5, "price": 200},
   		        {"product_id": 3, "quantity": 2, "price": 250}
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
			req, err := http.NewRequest("POST", "/api/v1/report/sale", bytes.NewBufferString(tt.rawJSON))
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
				mockRepo.AssertCalled(t, "InsertData", mock.Anything, mock.AnythingOfType("*models.CarriageReport"))
			} else {
				mockRepo.AssertNotCalled(t, "InsertData", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestInsertSalesEndpoint_InvalidRequestType(t *testing.T) {
	// Create a mock repository and a service based on it.
	mockRepo := &MockSalesRepository{}
	svc := service.NewSalesService(mockRepo)

	// Build the InsertSales endpoint.
	endpoint := httphandler.MakeInsertSalesEndpoint(svc)

	// Call the endpoint with an invalid request type (a string instead of *models.CarriageReport).
	resp, err := endpoint(context.Background(), "this is not a carriage")

	// Expect a nil response and an error indicating "invalid request type".
	assert.Nil(t, resp)
	assert.EqualError(t, err, "invalid request type")

	// Assert that InsertData was never called.
	mockRepo.AssertNotCalled(t, "InsertData", mock.Anything, mock.Anything)
}

// TestGetEmployeeCartsInTripEndpoint tests the GET /api/v1/report/sale/trip/cart/employee endpoint.
func TestGetEmployeeCartsInTripEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		mockSetup      func(m *MockSalesRepository)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Successful Get",
			queryParams: map[string]string{
				"route_id":    "route_test",
				"year":        "2023",
				"start_time":  "2023-01-15T10:00:01Z",
				"employee_id": "emp1",
			},
			mockSetup: func(m *MockSalesRepository) {
				sampleCart := models.Cart{
					CartID: models.CartID{
						EmployeeID:    "emp1",
						OperationTime: time.Date(2023, 1, 15, 12, 30, 0, 0, time.UTC),
					},
					OperationType: 1,
					Items: []models.Item{
						{ProductID: 1, Quantity: 10, Price: 100},
						{ProductID: 2, Quantity: 5, Price: 200},
					},
				}
				m.On("GetEmployeeCartsInTrip", mock.Anything, mock.AnythingOfType("*models.TripID"), mock.AnythingOfType("*string")).
					Return([]models.Cart{sampleCart}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: schemas.GetEmployeeCartsInTripResponse{
				Carts: []schemas.Cart{
					{
						CartID: schemas.CartID{
							EmployeeID:    "emp1",
							OperationTime: "2023-01-15T12:30:00Z",
						},
						OperationType: 1,
						Items: []schemas.Item{
							{ProductID: 1, Quantity: 10, Price: 100},
							{ProductID: 2, Quantity: 5, Price: 200},
						},
					},
				},
			},
		},
		{
			name: "Missing Query Parameters",
			queryParams: map[string]string{
				"route_id":   "route_test",
				"year":       "2023",
				"start_time": "2023-01-15T10:00:01Z",
				// "employee_id" is missing
			},
			mockSetup:      func(m *MockSalesRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: schemas.ErrorResponse{
				Error: "missing one or more required query parameters: route_id, year, start_time, employee_id",
			},
		},
		{
			name: "Invalid StartTime Format",
			queryParams: map[string]string{
				"route_id":    "route_test",
				"year":        "2023",
				"start_time":  "invalid-time",
				"employee_id": "emp1",
			},
			mockSetup:      func(m *MockSalesRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: schemas.ErrorResponse{
				Error: "invalid start_time format; must be RFC3339",
			},
		},
		{
			name: "Repository Error",
			queryParams: map[string]string{
				"route_id":    "route_test",
				"year":        "2023",
				"start_time":  "2023-01-15T10:00:01Z",
				"employee_id": "emp1",
			},
			mockSetup: func(m *MockSalesRepository) {
				m.On("GetEmployeeCartsInTrip", mock.Anything, mock.AnythingOfType("*models.TripID"), mock.AnythingOfType("*string")).
					Return(nil, errors.New("database error"))
			},
			// Even though repository returns an error, the endpoint still calls the repository because the query is valid.
			// With the current error encoder, the status code is returned as BadRequest.
			expectedStatus: http.StatusBadRequest,
			expectedBody: schemas.ErrorResponse{
				Error: "database error",
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

			// Create GET request for the endpoint
			req, err := http.NewRequest("GET", "/api/v1/report/trip/cart/employee", nil)
			assert.NoError(t, err, "Failed to create new GET request")

			// Set query parameters on the URL
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Set(key, value)
			}
			req.URL.RawQuery = q.Encode()

			// Create ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Serve the request using the HTTP handler
			handler.ServeHTTP(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code, "Unexpected status code")

			// Read and check response body
			body, err := io.ReadAll(rr.Body)
			assert.NoError(t, err, "Failed to read response body")

			expectedBodyJSON, _ := json.Marshal(tt.expectedBody)
			assert.JSONEq(t, string(expectedBodyJSON), string(body), "Response body does not match")

			// Determine if repository call is expected:
			// Repository should be called if both "employee_id" exists and "start_time" is valid.
			employeeID, hasEmployeeID := tt.queryParams["employee_id"]
			startTimeStr, hasStartTime := tt.queryParams["start_time"]
			_, validStartTime := time.Parse(time.RFC3339, startTimeStr)

			if hasEmployeeID && employeeID != "" && hasStartTime && validStartTime == nil {
				mockRepo.AssertCalled(t, "GetEmployeeCartsInTrip", mock.Anything, mock.AnythingOfType("*models.TripID"), mock.AnythingOfType("*string"))
			} else {
				mockRepo.AssertNotCalled(t, "GetEmployeeCartsInTrip", mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}

func TestGetEmployeeCartsInTripEndpoint_InvalidRequestType(t *testing.T) {
	// Create a mock repository and service.
	mockRepo := &MockSalesRepository{}
	svc := service.NewSalesService(mockRepo)

	// Build the GetEmployeeCartsInTrip endpoint.
	endpoint := httphandler.MakeGetEmployeeCartsInTripEndpoint(svc)

	// Call the endpoint directly with an invalid request type (e.g. a string)
	resp, err := endpoint(context.Background(), "this is not a valid GetEmployeeCartsInTrip request")

	// Assert that no response is returned and the error matches our expected error.
	assert.Nil(t, resp)
	assert.EqualError(t, err, "invalid request type")

	// Since the type assertion fails, the service method should never be called.
	mockRepo.AssertNotCalled(t, "GetEmployeeCartsInTrip", mock.Anything, mock.Anything, mock.Anything)
}

func TestGetEmployeeIDsByTripEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		mockSetup      func(m *MockSalesRepository)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Successful Get",
			queryParams: map[string]string{
				"route_id":   "route_test",
				"year":       "2023",
				"start_time": "2023-01-15T10:00:01Z",
			},
			mockSetup: func(m *MockSalesRepository) {
				// When called with a valid trip, return a list of employee IDs.
				m.On("GetEmployeeIDsByTrip", mock.Anything, mock.AnythingOfType("*models.TripID")).
					Return([]string{"emp1", "emp2"}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: schemas.GetEmployeeIDsByTripResponse{
				EmployeeIDs: []string{"emp1", "emp2"},
			},
		},
		{
			name: "Missing Query Parameters",
			queryParams: map[string]string{
				// "route_id" is missing in this case
				"year":       "2023",
				"start_time": "2023-01-15T10:00:01Z",
			},
			mockSetup:      func(m *MockSalesRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: schemas.ErrorResponse{
				Error: "missing required query parameters: route_id, year or start_time",
			},
		},
		{
			name: "Invalid StartTime Format",
			queryParams: map[string]string{
				"route_id":   "route_test",
				"year":       "2023",
				"start_time": "invalid-time",
			},
			mockSetup:      func(m *MockSalesRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: schemas.ErrorResponse{
				Error: "invalid start_time format; must be RFC3339",
			},
		},
		{
			name: "Repository Error",
			queryParams: map[string]string{
				"route_id":   "route_test",
				"year":       "2023",
				"start_time": "2023-01-15T10:00:01Z",
			},
			mockSetup: func(m *MockSalesRepository) {
				m.On("GetEmployeeIDsByTrip", mock.Anything, mock.AnythingOfType("*models.TripID")).
					Return(nil, errors.New("database error"))
			},
			// Given our error encoder (which always treats errors as validation errors),
			// the status code will be BadRequest.
			expectedStatus: http.StatusBadRequest,
			expectedBody: schemas.ErrorResponse{
				Error: "database error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize the mock repository.
			mockRepo := &MockSalesRepository{}
			tt.mockSetup(mockRepo)

			// Initialize the service with the mock repository.
			svc := service.NewSalesService(mockRepo)

			// Create the HTTP handler with the service.
			handler := httphandler.NewHTTPHandler(svc, log.NewNopLogger())

			// Create the GET request for the endpoint.
			req, err := http.NewRequest("GET", "/api/v1/report/trip/employee_id", nil)
			assert.NoError(t, err, "Failed to create new GET request")

			// Set query parameters on the URL.
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Set(key, value)
			}
			req.URL.RawQuery = q.Encode()

			// Create a ResponseRecorder to record the response.
			rr := httptest.NewRecorder()

			// Serve the request using the HTTP handler.
			handler.ServeHTTP(rr, req)

			// Check the status code.
			assert.Equal(t, tt.expectedStatus, rr.Code, "Unexpected status code")

			// Read and check the response body.
			body, err := io.ReadAll(rr.Body)
			assert.NoError(t, err, "Failed to read response body")

			expectedBodyJSON, _ := json.Marshal(tt.expectedBody)
			assert.JSONEq(t, string(expectedBodyJSON), string(body), "Response body does not match")

			// Repository should be called if both required query parameters are valid.
			routeID, hasRouteID := tt.queryParams["route_id"]
			startTimeStr, hasStartTime := tt.queryParams["start_time"]
			_, validStartTime := time.Parse(time.RFC3339, startTimeStr)

			if hasRouteID && routeID != "" && hasStartTime && validStartTime == nil {
				mockRepo.AssertCalled(t, "GetEmployeeIDsByTrip", mock.Anything, mock.AnythingOfType("*models.TripID"))
			} else {
				mockRepo.AssertNotCalled(t, "GetEmployeeIDsByTrip", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestGetEmployeeIDsByTripEndpoint_InvalidRequestType(t *testing.T) {
	// Create a mock repository and service.
	mockRepo := &MockSalesRepository{}
	svc := service.NewSalesService(mockRepo)

	// Build the GetEmployeeIDsByTrip endpoint.
	endpoint := httphandler.MakeGetEmployeeIDsByTripEndpoint(svc)

	// Call the endpoint directly with an invalid request type (e.g., a string).
	resp, err := endpoint(context.Background(), "this is not a valid GetEmployeeIDsByTrip request")

	// Expect a nil response and an error indicating "invalid request type".
	assert.Nil(t, resp)
	assert.EqualError(t, err, "invalid request type")

	// Assert that the repository's GetEmployeeIDsByTrip method was never called.
	mockRepo.AssertNotCalled(t, "GetEmployeeIDsByTrip", mock.Anything, mock.Anything)
}

func TestGetEmployeeTripsEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		mockSetup      func(m *MockSalesRepository)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Successful Get",
			queryParams: map[string]string{
				"employee_id": "emp1",
				"year":        "2023",
			},
			mockSetup: func(m *MockSalesRepository) {
				// Create a sample trip returned by the repository.
				sampleTrip := models.EmployeeTrip{
					EmployeeID: "emp1",
					TripID: models.TripID{
						RouteID:   "route_test",
						Year:      "2023",
						StartTime: time.Date(2023, 1, 15, 10, 0, 1, 0, time.UTC),
					},
					EndTime: time.Date(2023, 1, 15, 11, 0, 1, 0, time.UTC),
				}
				m.On("GetEmployeeTrips", mock.Anything, "emp1", "2023").
					Return([]models.EmployeeTrip{sampleTrip}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: schemas.GetEmployeeTripsResponse{
				EmployeeTrips: []schemas.EmployeeTrip{
					{
						EmployeeID: "emp1",
						TripID: schemas.TripID{
							RouteID:   "route_test",
							Year:      "2023",
							StartTime: "2023-01-15T10:00:01Z",
						},
						// EndTime remains as a time.Time in the schema.
						EndTime: time.Date(2023, 1, 15, 11, 0, 1, 0, time.UTC),
					},
				},
			},
		},
		{
			name: "Missing Query Parameters",
			queryParams: map[string]string{
				"year": "2023", // Missing employee_id.
			},
			mockSetup:      func(m *MockSalesRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: schemas.ErrorResponse{
				Error: "missing required query parameters: employee_id or year",
			},
		},
		{
			name: "Repository Error",
			queryParams: map[string]string{
				"employee_id": "emp1",
				"year":        "2023",
			},
			mockSetup: func(m *MockSalesRepository) {
				m.On("GetEmployeeTrips", mock.Anything, "emp1", "2023").
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: schemas.ErrorResponse{
				Error: "database error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize the mock repository and set up expectations.
			mockRepo := &MockSalesRepository{}
			tt.mockSetup(mockRepo)

			// Create the service and HTTP handler.
			svc := service.NewSalesService(mockRepo)
			handler := httphandler.NewHTTPHandler(svc, log.NewNopLogger())

			// Create a new GET request.
			req, err := http.NewRequest("GET", "/api/v1/report/trip/employee_trip", nil)
			assert.NoError(t, err, "Failed to create new GET request")

			// Set the query parameters.
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Set(key, value)
			}
			req.URL.RawQuery = q.Encode()

			// Create a ResponseRecorder to capture the response.
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Assert that the returned status code is as expected.
			assert.Equal(t, tt.expectedStatus, rr.Code, "Unexpected status code")

			// Read and compare the response body.
			body, err := io.ReadAll(rr.Body)
			assert.NoError(t, err, "Failed to read response body")
			expectedBodyJSON, _ := json.Marshal(tt.expectedBody)
			assert.JSONEq(t, string(expectedBodyJSON), string(body), "Response body does not match")

			// Determine if the repository should have been called.
			empID, hasEmp := tt.queryParams["employee_id"]
			year, hasYear := tt.queryParams["year"]

			if hasEmp && empID != "" && hasYear && year != "" {
				mockRepo.AssertCalled(t, "GetEmployeeTrips", mock.Anything, empID, year)
			} else {
				mockRepo.AssertNotCalled(t, "GetEmployeeTrips", mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}

func TestGetEmployeeTripsEndpoint_InvalidRequestType(t *testing.T) {
	// Create a mock repository and service.
	mockRepo := &MockSalesRepository{}
	svc := service.NewSalesService(mockRepo)

	// Build the GetEmployeeTrips endpoint.
	endpoint := httphandler.MakeGetEmployeeTripsEndpoint(svc)

	// Call the endpoint directly with an invalid request type (e.g., a string).
	resp, err := endpoint(context.Background(), "this is not a valid GetEmployeeTrips request")

	// Expect a nil response and an error indicating "invalid request type".
	assert.Nil(t, resp)
	assert.EqualError(t, err, "invalid request type")

	// Assert that the repository's GetEmployeeTrips method was never called.
	mockRepo.AssertNotCalled(t, "GetEmployeeTrips", mock.Anything, mock.Anything, mock.Anything)
}

// TestUpdateItemQuantityEndpoint tests the PUT /api/v1/report/trip/cart/item/quantity endpoint.
func TestUpdateItemQuantityEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		rawJSON        string
		mockSetup      func(m *MockSalesRepository)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Successful Update",
			rawJSON: `{
				"trip_id": {
					"route_id": "route_test",
					"start_time": "2023-01-15T10:00:01Z"
				},
				"cart_id": {
					"employee_id": "emp1",
					"operation_time": "2023-01-15T12:30:00Z"
				},
				"product_id": 1,
				"new_quantity": 15
			}`,
			mockSetup: func(m *MockSalesRepository) {
				// Expect a call with a valid TripID, CartID, productID and newQuantity.
				m.On("UpdateItemQuantity", mock.Anything, mock.AnythingOfType("*models.TripID"),
					mock.AnythingOfType("*models.CartID"), mock.AnythingOfType("*int"), mock.AnythingOfType("*int16")).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: schemas.UpdateItemQuantityResponse{
				Message: "Item quantity updated successfully",
			},
		},
		{
			name: "Malformed JSON",
			rawJSON: `{
				"trip_id": {
					"route_id": "route_test",
					"start_time": "2023-01-15T10:00:01Z"
				,
				"cart_id": {
					"employee_id": "emp1",
					"operation_time": "2023-01-15T12:30:00Z"
				},
				"product_id": 1,
				"new_quantity": 15
			}`, // note the missing closing brace or extra comma making the JSON invalid.
			mockSetup:      func(m *MockSalesRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: schemas.ErrorResponse{
				Error: "invalid request body",
			},
		},
		{
			name: "Invalid StartTime Format",
			rawJSON: `{
				"trip_id": {
					"route_id": "route_test",
					"start_time": "not-a-time"
				},
				"cart_id": {
					"employee_id": "emp1",
					"operation_time": "2023-01-15T12:30:00Z"
				},
				"product_id": 1,
				"new_quantity": 15
			}`,
			mockSetup:      func(m *MockSalesRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: schemas.ErrorResponse{
				Error: "invalid start_time format; must be RFC3339",
			},
		},
		{
			name: "Invalid OperationTime Format",
			rawJSON: `{
				"trip_id": {
					"route_id": "route_test",
					"start_time": "2023-01-15T10:00:01Z"
				},
				"cart_id": {
					"employee_id": "emp1",
					"operation_time": "bad-time"
				},
				"product_id": 1,
				"new_quantity": 15
			}`,
			mockSetup:      func(m *MockSalesRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: schemas.ErrorResponse{
				Error: "invalid operation_time format; must be RFC3339",
			},
		},
		{
			name: "Repository Error",
			rawJSON: `{
				"trip_id": {
					"route_id": "route_test",
					"start_time": "2023-01-15T10:00:01Z"
				},
				"cart_id": {
					"employee_id": "emp1",
					"operation_time": "2023-01-15T12:30:00Z"
				},
				"product_id": 1,
				"new_quantity": 15
			}`,
			mockSetup: func(m *MockSalesRepository) {
				m.On("UpdateItemQuantity", mock.Anything, mock.AnythingOfType("*models.TripID"),
					mock.AnythingOfType("*models.CartID"), mock.AnythingOfType("*int"), mock.AnythingOfType("*int16")).
					Return(errors.New("update failed"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: schemas.ErrorResponse{
				Error: "update failed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the mock repository.
			mockRepo := &MockSalesRepository{}
			tt.mockSetup(mockRepo)

			// Create the service and HTTP handler.
			svc := service.NewSalesService(mockRepo)
			handler := httphandler.NewHTTPHandler(svc, log.NewNopLogger())

			// Create the PUT request with the test JSON.
			req, err := http.NewRequest("PUT", "/api/v1/report/trip/cart/item/quantity", bytes.NewBufferString(tt.rawJSON))
			assert.NoError(t, err, "Failed to create new PUT request")
			req.Header.Set("Content-Type", "application/json")

			// Create a ResponseRecorder.
			rr := httptest.NewRecorder()

			// Serve the request.
			handler.ServeHTTP(rr, req)

			// Assert on the response status code.
			assert.Equal(t, tt.expectedStatus, rr.Code, "Unexpected status code")

			// Read the response body.
			body, err := io.ReadAll(rr.Body)
			assert.NoError(t, err, "Failed to read response body")

			expectedBodyJSON, _ := json.Marshal(tt.expectedBody)
			assert.JSONEq(t, string(expectedBodyJSON), string(body), "Response body does not match")

			// Determine if the repository should be called:
			// Repository is expected if the JSON is well-formed and both time fields parse correctly.
			var reqBody schemas.UpdateItemQuantityRequest
			parseErr := json.Unmarshal([]byte(tt.rawJSON), &reqBody)
			_, tripTimeErr := time.Parse(time.RFC3339, reqBody.TripID.StartTime)
			_, cartTimeErr := time.Parse(time.RFC3339, reqBody.CartID.OperationTime)

			if parseErr == nil && tripTimeErr == nil && cartTimeErr == nil {
				mockRepo.AssertCalled(t, "UpdateItemQuantity", mock.Anything, mock.AnythingOfType("*models.TripID"),
					mock.AnythingOfType("*models.CartID"), mock.AnythingOfType("*int"), mock.AnythingOfType("*int16"))
			} else {
				mockRepo.AssertNotCalled(t, "UpdateItemQuantity", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}

func TestUpdateItemQuantityEndpoint_InvalidRequestType(t *testing.T) {
	// Create a mock repository and service.
	mockRepo := &MockSalesRepository{}
	svc := service.NewSalesService(mockRepo)

	// Build the UpdateItemQuantity endpoint.
	endpoint := httphandler.MakeUpdateItemQuantityEndpoint(svc)

	// Call the endpoint directly with an invalid request type (e.g., a string).
	resp, err := endpoint(context.Background(), "this is not a valid UpdateItemQuantity request")

	// Expect a nil response and an error indicating "invalid request type".
	assert.Nil(t, resp)
	assert.EqualError(t, err, "invalid request type")

	// Ensure the repository method is never called.
	mockRepo.AssertNotCalled(t, "UpdateItemQuantity", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

// TestDeleteItemFromCartEndpoint tests the DELETE /api/v1/report/trip/cart/item endpoint.
func TestDeleteItemFromCartEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		rawJSON        string
		mockSetup      func(m *MockSalesRepository)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Successful Delete",
			rawJSON: `{
				"trip_id": {
					"route_id": "route_test",
					"start_time": "2023-01-15T10:00:01Z"
				},
				"cart_id": {
					"employee_id": "emp1",
					"operation_time": "2023-01-15T12:30:00Z"
				},
				"product_id": 1
			}`,
			mockSetup: func(m *MockSalesRepository) {
				m.On("DeleteItemFromCart", mock.Anything, mock.AnythingOfType("*models.TripID"),
					mock.AnythingOfType("*models.CartID"), mock.AnythingOfType("*int")).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: schemas.DeleteItemFromCartResponse{
				Message: "Item deleted successfully",
			},
		},
		{
			name: "Malformed JSON",
			rawJSON: `{
				"trip_id": {
					"route_id": "route_test",
					"start_time": "2023-01-15T10:00:01Z"
				,
				"cart_id": {
					"employee_id": "emp1",
					"operation_time": "2023-01-15T12:30:00Z"
				},
				"product_id": 1
			}`, // Malformed due to a misplaced comma or missing brace.
			mockSetup:      func(m *MockSalesRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: schemas.ErrorResponse{
				Error: "invalid request body",
			},
		},
		{
			name: "Invalid StartTime Format",
			rawJSON: `{
				"trip_id": {
					"route_id": "route_test",
					"start_time": "not-a-time"
				},
				"cart_id": {
					"employee_id": "emp1",
					"operation_time": "2023-01-15T12:30:00Z"
				},
				"product_id": 1
			}`,
			mockSetup:      func(m *MockSalesRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: schemas.ErrorResponse{
				Error: "invalid start_time format; must be RFC3339",
			},
		},
		{
			name: "Invalid OperationTime Format",
			rawJSON: `{
				"trip_id": {
					"route_id": "route_test",
					"start_time": "2023-01-15T10:00:01Z"
				},
				"cart_id": {
					"employee_id": "emp1",
					"operation_time": "bad-time"
				},
				"product_id": 1
			}`,
			mockSetup:      func(m *MockSalesRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: schemas.ErrorResponse{
				Error: "invalid operation_time format; must be RFC3339",
			},
		},
		{
			name: "Repository Error",
			rawJSON: `{
				"trip_id": {
					"route_id": "route_test",
					"start_time": "2023-01-15T10:00:01Z"
				},
				"cart_id": {
					"employee_id": "emp1",
					"operation_time": "2023-01-15T12:30:00Z"
				},
				"product_id": 1
			}`,
			mockSetup: func(m *MockSalesRepository) {
				m.On("DeleteItemFromCart", mock.Anything, mock.AnythingOfType("*models.TripID"),
					mock.AnythingOfType("*models.CartID"), mock.AnythingOfType("*int")).
					Return(errors.New("item does not exist"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: schemas.ErrorResponse{
				Error: "item does not exist",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the mock repository.
			mockRepo := &MockSalesRepository{}
			tt.mockSetup(mockRepo)

			// Create the service and HTTP handler.
			svc := service.NewSalesService(mockRepo)
			handler := httphandler.NewHTTPHandler(svc, log.NewNopLogger())

			// Create the DELETE request with the test JSON.
			req, err := http.NewRequest("DELETE", "/api/v1/report/trip/cart/item", bytes.NewBufferString(tt.rawJSON))
			assert.NoError(t, err, "Failed to create new DELETE request")
			req.Header.Set("Content-Type", "application/json")

			// Create a ResponseRecorder.
			rr := httptest.NewRecorder()

			// Serve the request.
			handler.ServeHTTP(rr, req)

			// Assert on the response status code.
			assert.Equal(t, tt.expectedStatus, rr.Code, "Unexpected status code")

			// Read the response body.
			body, err := io.ReadAll(rr.Body)
			assert.NoError(t, err, "Failed to read response body")

			expectedBodyJSON, _ := json.Marshal(tt.expectedBody)
			assert.JSONEq(t, string(expectedBodyJSON), string(body), "Response body does not match")

			// Determine if the repository should be called.
			// We try to unmarshal the JSON into the expected request structure and parse the time fields.
			var reqBody schemas.DeleteItemFromCartRequest
			parseErr := json.Unmarshal([]byte(tt.rawJSON), &reqBody)
			_, tripTimeErr := time.Parse(time.RFC3339, reqBody.TripID.StartTime)
			_, cartTimeErr := time.Parse(time.RFC3339, reqBody.CartID.OperationTime)

			if parseErr == nil && tripTimeErr == nil && cartTimeErr == nil {
				mockRepo.AssertCalled(t, "DeleteItemFromCart", mock.Anything, mock.AnythingOfType("*models.TripID"),
					mock.AnythingOfType("*models.CartID"), mock.AnythingOfType("*int"))
			} else {
				mockRepo.AssertNotCalled(t, "DeleteItemFromCart", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}

func TestDeleteItemFromCartEndpoint_InvalidRequestType(t *testing.T) {
	// Create a mock repository and service.
	mockRepo := &MockSalesRepository{}
	svc := service.NewSalesService(mockRepo)

	// Build the DeleteItemFromCart endpoint.
	endpoint := httphandler.MakeDeleteItemFromCartEndpoint(svc)

	// Call the endpoint directly with an invalid request type (e.g., a string).
	resp, err := endpoint(context.Background(), "this is not a valid DeleteItemFromCart request")

	// Expect a nil response and an error indicating "invalid request type".
	assert.Nil(t, resp)
	assert.EqualError(t, err, "invalid request type")

	// Ensure the repository's DeleteItemFromCart method is never called.
	mockRepo.AssertNotCalled(t, "DeleteItemFromCart", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
