package repository

import (
	"ChaikaReports/internal/models"
	"github.com/gocql/gocql"
)

type SalesRepository interface {
	InsertData(salesData *models.SalesData) error
	GetActionsByConductor(routeID, tripID, conductorID int) ([]models.Action, error)
	UpdateActionCount(operationID gocql.UUID, routeID, tripID, productID, newCount int) error
	DeleteProductFromAction(operationID gocql.UUID, routeID, tripID, conductorID, productID int) error
	GetConductorsByTripID(routeID, tripID int) ([]models.SalesData, error)
}
