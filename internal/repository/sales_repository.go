package repository

import "ChaikaReports/internal/models"

type SalesRepository interface {
	// Inserts all data from a Carriage into the Cassandra database
	InsertData(carriageReport *models.Carriage) error

	// Gets all carts employee has sold during trip, returns array of Carts
	GetEmployeeCartsInTrip(tripID *models.TripID, employeeID *string) ([]models.Cart, error)

	// Gets all employees in trip
	GetEmployeeIDsByTrip(tripID *models.TripID) ([]string, error)

	//Updates item quanity in cart
	UpdateItemQuantity(tripID *models.TripID, cartID *models.CartID, productID int, newQuantity *int16) error

	//Delete item from cart
	DeleteItemFromCart(tripID *models.TripID, cartID *models.CartID, productID int) error
}
