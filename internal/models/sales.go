package models

type Action struct {
	ProductID       int `json:"productID"`
	OperationTypeID int `json:"operationTypeID"`
	Count           int `json:"count"`
}

type SalesData struct {
	TripID      int      `json:"tripID"`
	CarriageID  int      `json:"carriageID"`
	ConductorID int      `json:"conductorID"`
	Actions     []Action `json:"actions"`
}
