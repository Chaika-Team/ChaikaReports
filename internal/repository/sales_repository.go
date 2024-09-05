package repository

import (
	"ChaikaReports/internal/models"
	"github.com/gocql/gocql"
)

type SalesRepository interface {
	// InsertData Inserts all data from a Carriage into the Cassandra database
	InsertData(session *gocql.Session, carriageReport *models.Carriage) error

	// GetEmployeeCartsInTrip Gets all carts employee has sold during trip, returns array of Carts
	GetEmployeeCartsInTrip(session *gocql.Session, tripID *models.TripID, employeeID *string) ([]models.Cart, error)

	// GetEmployeeIDsByTrip Gets all employees in trip
	GetEmployeeIDsByTrip(session *gocql.Session, tripID *models.TripID) ([]string, error)

	// UpdateItemQuantity Updates item quantity in cart
	UpdateItemQuantity(session *gocql.Session, tripID *models.TripID, cartID *models.CartID, productID int, newQuantity *int16) error

	// DeleteItemFromCart Deletes item from cart
	DeleteItemFromCart(session *gocql.Session, tripID *models.TripID, cartID *models.CartID, productID int) error
}
