package schemas

// TripID represents the trip identifier in the request
type TripID struct {
	RouteID   string `json:"route_id" validate:"required"`
	StartTime string `json:"start_time" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
}

// CartSchema represents each cart in the request
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

// ItemSchema represents each item in the cart
type Item struct {
	ProductID int   `json:"product_id" validate:"required"`
	Quantity  int16 `json:"quantity" validate:"required"`
	Price     int64 `json:"price" validate:"required,min=0"` //Storing price in kopeeks
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

// ErrorResponse represents the error response body
type ErrorResponse struct {
	Error string `json:"error"`
}
