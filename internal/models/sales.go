package models

import "time"

type Action struct {
	ProductID       int       `json:"productID"`
	OperationTypeID int       `json:"operationTypeID"`
	OperationTime   time.Time `json:"OperationTime"`
	RouteID         int       `json:"routeID"`
	Count           int       `json:"count"`
}

type SalesData struct {
	TripID      int      `json:"tripID"`
	CarriageID  int      `json:"carriageID"`
	ConductorID int      `json:"conductorID"`
	Actions     []Action `json:"actions"`
}
