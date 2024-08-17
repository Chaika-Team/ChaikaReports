package repository

import (
	"ChaikaReports/internal/models"
)

type SalesRepository interface {
	InsertData(salesData *models.SalesData) error
	GetActionsByConductor(salesData *models.SalesData) ([]models.Action, error)
	UpdateActionCount(salesData *models.SalesData, action *models.Action) error
	DeleteProductFromAction(salesData *models.SalesData, action *models.Action) error
	GetConductorsByTripID(salesData *models.SalesData) ([]models.SalesData, error)
}
