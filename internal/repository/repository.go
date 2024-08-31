package repository

import "ChaikaReports/internal/models"

type SalesRepository interface {
	InsertData(operations models.Operations) error
	GetConductorCarts()
	UpdateItemCount() error
	DeleteItemFromCart() error
	GetEmployeesByTrip()
}
