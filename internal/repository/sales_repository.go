package repository

import (
	"ChaikaReports/internal/models"
	"context"
)

type SalesRepository interface {
	// InsertData Inserts all data from a Carriage into the Cassandra database
	InsertData(ctx context.Context, carriageReport *models.Carriage) error

	// GetEmployeeCartsInTrip Gets all carts employee has sold during trip, returns array of Carts
	GetEmployeeCartsInTrip(ctx context.Context, tripID *models.TripID, employeeID *string) ([]models.Cart, error)

	// GetEmployeeIDsByTrip Gets all employees in trip
	GetEmployeeIDsByTrip(ctx context.Context, tripID *models.TripID) ([]string, error)

	// UpdateItemQuantity Updates item quantity in cart
	UpdateItemQuantity(ctx context.Context, tripID *models.TripID, cartID *models.CartID, productID *int, newQuantity *int16) error

	// DeleteItemFromCart Deletes item from cart
	DeleteItemFromCart(ctx context.Context, tripID *models.TripID, cartID *models.CartID, productID *int) error
}
