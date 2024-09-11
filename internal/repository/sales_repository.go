package repository

import (
	"ChaikaReports/internal/models"
)

type SalesRepository interface {
	// InsertData Inserts all data from a Carriage into the Cassandra database
	InsertData(carriageReport *models.Carriage) error

	// GetEmployeeCartsInTrip Gets all carts employee has sold during trip, returns array of Carts
	GetEmployeeCartsInTrip(tripID *models.TripID, employeeID *string) ([]models.Cart, error)

	// GetEmployeeIDsByTrip Gets all employees in trip
	GetEmployeeIDsByTrip(tripID *models.TripID) ([]string, error)

	// UpdateItemQuantity Updates item quantity in cart
	UpdateItemQuantity(tripID *models.TripID, cartID *models.CartID, productID int, newQuantity *int16) error

	// DeleteItemFromCart Deletes item from cart
	DeleteItemFromCart(tripID *models.TripID, cartID *models.CartID, productID int) error
}
