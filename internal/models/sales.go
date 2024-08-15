package models

import (
	"github.com/gocql/gocql"
	"time"
)

type Action struct {
	ProductID       int        `json:"productID"`
	OperationTypeID int        `json:"operationTypeID"`
	OperationTime   time.Time  `json:"operationTime"`
	OperationAmount float64    `json:"operationAmount"`
	OperationID     gocql.UUID `json:"operationID"`
	Count           int        `json:"count"`
}

type SalesData struct {
	RouteID     int      `json:"routeID"`
	TripID      int      `json:"tripID"`
	CarriageID  int      `json:"carriageID"`
	ConductorID int      `json:"conductorID"`
	Actions     []Action `json:"actions"`
}
