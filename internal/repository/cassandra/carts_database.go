package cassandra

import (
	"ChaikaReports/internal/models"
	"ChaikaReports/internal/repository"
	"fmt"
	"github.com/gocql/gocql"
	"time"
)

type SalesRepository struct{}

func NewSalesRepository() repository.SalesRepository {
	return &SalesRepository{}
}

// Inserts all data from a Carriage into the Cassandra database
func (r *SalesRepository) InsertData(carriageReport *models.Carriage) error {
	for _, cart := range carriageReport.Carts {
		for _, item := range cart.Items {
			err := Session.Query(`
				INSERT INTO operations (route_id, start_time, end_time, carriage_num, employee_id, operation_type, operation_time, product_id, quantity, price)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				&carriageReport.TripID.RouteID,
				&carriageReport.TripID.StartTime,
				&carriageReport.EndTime,
				&carriageReport.CarriageNum,
				&cart.CartID.EmployeeID,
				&cart.OperationType,
				&cart.CartID.OperationTime,
				&item.ProductID,
				&item.Quantity,
				&item.Price,
			).Exec()

			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Gets all carts employee has sold during trip, returns array of Carts
func (r *SalesRepository) GetEmployeeCartsInTrip(tripID *models.TripID, employeeID *string) ([]models.Cart, error) {
	iter := Session.Query(`SELECT operation_time, operation_type, product_id, quantity, price FROM operations WHERE route_id = ? AND start_time = ? AND employee_id = ?`,
		&tripID.RouteID, &tripID.StartTime, &employeeID).Iter()
	// Uses helper function to convert query result into slice of models.Cart
	var carts, _ = aggregateCartsFromRows(iter, *employeeID) //TODO add error handling and logging

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return carts, nil
}

// Gets all employees by TripID (RouteID, StartTime)
func (r *SalesRepository) GetEmployeeIDsByTrip(tripID *models.TripID) ([]string, error) {
	iter := Session.Query(`SELECT employee_id FROM operations WHERE route_id = ? AND start_time = ?`,
		&tripID.RouteID, &tripID.StartTime).Iter()

	var employeeIDs []string
	var employeeID string
	for iter.Scan(&employeeID) {
		employeeIDs = append(employeeIDs, employeeID)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return employeeIDs, nil
}

// Updates quantity of items in cart
func (r *SalesRepository) UpdateItemQuantity(tripID *models.TripID, cartID *models.CartID, productID int, newQuantity *int16) error {
	return Session.Query(`UPDATE operations SET quantity = ? WHERE route_id = ? AND start_time = ? AND employee_id = ? AND operation_time = ? AND product_id = ?`,
		newQuantity, tripID.RouteID, tripID.StartTime, cartID.EmployeeID, cartID.OperationTime, productID).Exec()
}

// Deletes cart item (operation)
func (r *SalesRepository) DeleteItemFromCart(tripID *models.TripID, cartID *models.CartID, productID int) error {
	return Session.Query(`DELETE FROM operations WHERE route_id = ? AND start_time = ? AND employee_id = ? AND operation_time = ? AND product_id = ?`,
		tripID.RouteID, tripID.StartTime, cartID.EmployeeID, cartID.OperationTime, productID).Exec()
}

// Helper function to process rows and return an array of Carts
func aggregateCartsFromRows(iter *gocql.Iter, employeeID string) ([]models.Cart, error) {
	cartMap := make(map[string]*models.Cart)

	var operationTime time.Time
	var operationType int8
	var productID int
	var quantity int16
	var price float64

	for iter.Scan(&operationTime, &operationType, &productID, &quantity, &price) {

		// Create a cartID struct
		cartID := models.CartID{
			EmployeeID:    employeeID,
			OperationTime: operationTime,
		}

		//Declares key and ensures uniqueness by employeeID and operationTime
		cartKey := fmt.Sprintf("%s-%s", cartID.EmployeeID, cartID.OperationTime.Format(time.RFC3339))

		// Checks if cart key exists in current map and inserts items into corresponding key
		if cart, exists := cartMap[cartKey]; exists {
			cart.Items = append(cart.Items, models.Item{
				ProductID: productID,
				Quantity:  quantity,
				Price:     price,
			})
		} else { //Inserts new key and populates with items
			cartMap[cartKey] = &models.Cart{
				CartID:        cartID,
				OperationType: operationType,
				Items: []models.Item{
					{
						ProductID: productID,
						Quantity:  quantity,
						Price:     price,
					},
				},
			}
		}
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	// Convert map to slice of Carts
	carts := make([]models.Cart, 0, len(cartMap))
	for _, cart := range cartMap {
		carts = append(carts, *cart)
	}

	return carts, nil
}
