package repository

import "ChaikaReports/internal/models"

type SalesRepository interface {
	InsertData(report models.CarriageReport) error
	GetConductorCarts()
	UpdateItemCount() error
	DeleteItemFromCart() error
	GetEmployeesByTrip()
}
