package models

import "time"

// Operation types in Cart
const (
	OperationTypeSale   int8 = 1
	OperationTypeRefund int8 = 2
)

// Item is a domain model that specifies the quantity, id and price of a product in a cart
type Item struct {
	ProductID int   `json:"product_id"`
	Quantity  int16 `json:"quantity"`
	Price     int64 `json:"price"`
}

type Cart struct {
	CartID        CartID `json:"cart_id"`
	OperationType int8   `json:"operation_type"`
	Items         []Item `json:"items"`
}

type Carriage struct {
	TripID     TripID    `json:"trip_id"`
	EndTime    time.Time `json:"end_time"`
	CarriageID int8      `json:"carriage_id" validate:"required,gte=0,lte=127"`
	Carts      []Cart    `json:"carts"`
}

type TripID struct {
	RouteID   string    `json:"route_id"`
	Year      string    `json:"year"`
	StartTime time.Time `json:"start_time"`
}

type CartID struct {
	EmployeeID    string    `json:"employee_id"`
	OperationTime time.Time `json:"operation_time"`
}

// EmployeeTrip is a domain model that represents a trip that an employee was in
type EmployeeTrip struct {
	EmployeeID string    `json:"employee_id"`
	TripID     TripID    `json:"trip_id"`
	EndTime    time.Time `json:"end_time"`
}
