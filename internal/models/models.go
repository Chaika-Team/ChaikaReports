package models

import "time"

type Item struct {
	ProductID int     `json:"product_id"`
	Quantity  int16   `json:"quantity"`
	Price     float32 `json:"price"`
}

type Cart struct {
	CartID        CartID `json:"cart_id"`
	OperationType int8   `json:"operation_type"`
	Items         []Item `json:"items"`
}

type Carriage struct {
	TripID     TripID    `json:"trip_id"`
	EndTime    time.Time `json:"end_time"`
	CarriageID int8      `json:"carriage_id"`
	Carts      []Cart    `json:"carts"`
}

type TripID struct {
	RouteID   string    `json:"route_id"`
	StartTime time.Time `json:"start_time"`
}

type CartID struct {
	EmployeeID    string    `json:"employee_id"`
	OperationTime time.Time `json:"operation_time"`
}
