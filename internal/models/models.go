package models

import "time"

type Item struct {
	ProductID int   `json:"productID"`
	Quantity  int16 `json:"quantity"`
}

type Cart struct {
	EmployeeID    int       `json:"employeeID"`
	OperationType int8      `json:"operationType"`
	OperationTime time.Time `json:"operationTime"`
	Items         []Item
}

type Operation struct {
	RouteID     string    `json:"routeID"`
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime"`
	CarriageNum int8      `json:"carriageNum"`
	Carts       []Cart    `json:"carts"`
}
