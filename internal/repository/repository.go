package repository

import "ChaikaReports/internal/models"

type SalesRepository interface {
	InsertSalesData(salesData *models.SalesData) error
	GetActionsByConductor(tripID, conductorID int) ([]models.Action, error)
	UpdateActionCount(tripID, carriageID, conductorID, productID, operationTypeID, newCount int) error
	DeleteActions(tripID, conductorID int) error
	GetConductorsByTripID(tripID int) ([]models.SalesData, error)
}
