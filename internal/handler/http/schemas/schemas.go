package schemas

// InsertSalesRequest represents the request body for the POST /api/v1/sales endpoint
type InsertSalesRequest struct {
	TripID     TripID       `json:"trip_id" validate:"required"`
	EndTime    string       `json:"end_time" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	CarriageID int8         `json:"carriage_id" validate:"required"`
	Carts      []CartSchema `json:"carts" validate:"required,dive"`
}

// TripID represents the trip identifier in the request
type TripID struct {
	RouteID   string `json:"route_id" validate:"required"`
	StartTime string `json:"start_time" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
}

// CartSchema represents each cart in the request
type CartSchema struct {
	CartID        CartID       `json:"cart_id" validate:"required"`
	OperationType int8         `json:"operation_type" validate:"required"`
	Items         []ItemSchema `json:"items" validate:"required"`
}

// CartID represents the cart identifier in the request
type CartID struct {
	EmployeeID    string `json:"employee_id" validate:"required"`
	OperationTime string `json:"operation_time" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
}

// ItemSchema represents each item in the cart
type ItemSchema struct {
	ProductID int   `json:"product_id" validate:"required"`
	Quantity  int16 `json:"quantity" validate:"required"`
	Price     int64 `json:"price" validate:"required,min=0"` //Storing price in kopeeks
}

// InsertSalesResponse represents the response body for a successful insert
type InsertSalesResponse struct {
	Message string `json:"message"`
}

// ErrorResponse represents the error response body
type ErrorResponse struct {
	Error string `json:"error"`
}
