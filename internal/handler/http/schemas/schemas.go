package schemas

import "time"

// TripID represents the trip identifier in the request
type TripID struct {
	RouteID   string `json:"route_id" validate:"required"`
	StartTime string `json:"start_time" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
}

// Cart represents each cart in the request
type Cart struct {
	CartID        CartID `json:"cart_id" validate:"required"`
	OperationType int8   `json:"operation_type" validate:"required"`
	Items         []Item `json:"items" validate:"required"`
}

// CartID represents the cart identifier in the request
type CartID struct {
	EmployeeID    string `json:"employee_id" validate:"required"`
	OperationTime string `json:"operation_time" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
}

// Item represents each item in the cart
type Item struct {
	ProductID int   `json:"product_id" validate:"required"`
	Quantity  int16 `json:"quantity" validate:"required"`
	Price     int64 `json:"price" validate:"required,min=0"` //Storing price in kopeeks
}

type EmployeeTrip struct {
	EmployeeID string    `json:"employee_id"`
	Year       string    `json:"year"`
	TripID     TripID    `json:"trip_id"`
	EndTime    time.Time `json:"end_time"`
}

// InsertSalesRequest represents the request body for the POST /api/v1/sales endpoint
type InsertSalesRequest struct {
	TripID     TripID `json:"trip_id" validate:"required"`
	EndTime    string `json:"end_time" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	CarriageID int8   `json:"carriage_id" validate:"required"`
	Carts      []Cart `json:"carts" validate:"required,dive"`
}

// InsertSalesResponse represents the response body for a successful insert
type InsertSalesResponse struct {
	Message string `json:"message"`
}

// GetEmployeeCartsInTripRequest represents the request for the GET /api/v1/sales/trip/cart/employee endpoint.
type GetEmployeeCartsInTripRequest struct {
	TripID     TripID `json:"trip_id" validate:"required"`
	EmployeeID string `json:"employee_id" validate:"required"`
}

// GetEmployeeCartsInTripResponse represents the response with the list of carts.
type GetEmployeeCartsInTripResponse struct {
	Carts []Cart `json:"carts"`
}

type GetEmployeeIDsByTripRequest struct {
	TripID TripID `json:"trip_id" validate:"required"`
}

// GetEmployeeIDsByTripResponse represents the response with the list of employee IDs.
type GetEmployeeIDsByTripResponse struct {
	EmployeeIDs []string `json:"employee_ids"`
}

type GetEmployeeTripsRequest struct {
	EmployeeID string `json:"employee_id" validate:"required"`
	Year       string `json:"year" validate:"required"`
}

type GetEmployeeTripsResponse struct {
	EmployeeTrips []EmployeeTrip `json:"employee_trips" validate:"required"`
}

type UpdateItemQuantityRequest struct {
	TripID      TripID `json:"trip_id" validate:"required"`
	CartID      CartID `json:"cart_id" validate:"required"`
	ProductID   int    `json:"product_id" validate:"required"`
	NewQuantity int16  `json:"new_quantity" validate:"required"`
}

// UpdateItemQuantityResponse represents the response body for a successful update.
type UpdateItemQuantityResponse struct {
	Message string `json:"message"`
}

type DeleteItemFromCartRequest struct {
	TripID    TripID `json:"trip_id" validate:"required"`
	CartID    CartID `json:"cart_id" validate:"required"`
	ProductID int    `json:"product_id" validate:"required"`
}

// DeleteItemFromCartResponse represents the response body for a successful deletion.
type DeleteItemFromCartResponse struct {
	Message string `json:"message"`
}

// ErrorResponse represents the error response body
type ErrorResponse struct {
	Error string `json:"error"`
}
