package cassandra

import (
	"ChaikaReports/internal/models"
	"context"
	"fmt"
	"time"

	"github.com/go-kit/log"
	"github.com/gocql/gocql"
)

type SalesRepository struct {
	session *gocql.Session
	log     log.Logger
}

func NewSalesRepository(session *gocql.Session, logger log.Logger) *SalesRepository {
	return &SalesRepository{
		session: session,
		log:     logger,
	}

}

const insertOperationQuery = `
	INSERT INTO operations (route_id, start_time, end_time, carriage_id, 
	employee_id, operation_type, operation_time, product_id, quantity, price)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

const getEmployeeCartsInTripQuery = `SELECT operation_time, operation_type, product_id, quantity, price
	FROM operations
	WHERE route_id = ?
	  AND start_time = ?
	  AND employee_id = ?`

const getEmployeeIdsByTripQuery = `SELECT employee_id 
	FROM operations 
	WHERE route_id = ?
	  AND start_time = ?`

const updateItemQuantityQuery = `UPDATE operations SET quantity = ? 
    WHERE route_id = ? 
      AND start_time = ?
      AND employee_id = ?
      AND operation_time = ?
      AND product_id = ?
    IF EXISTS`

const deleteItemFromCartQuery = `DELETE FROM operations WHERE route_id = ? AND start_time = ? AND employee_id = ? AND operation_time = ? AND product_id = ?`

// InsertData Inserts all data from a Carriage into the Cassandra database
func (r *SalesRepository) InsertData(ctx context.Context, carriageReport *models.Carriage) error {
	batch := r.session.NewBatch(gocql.LoggedBatch).WithContext(ctx)
	for _, cart := range carriageReport.Carts {
		for _, item := range cart.Items {
			// Batch query allows to save data integrity by stopping transaction if at least one insertion fails
			batch.Query(insertOperationQuery,
				&carriageReport.TripID.RouteID,
				&carriageReport.TripID.StartTime,
				&carriageReport.EndTime,
				&carriageReport.CarriageID,
				&cart.CartID.EmployeeID,
				&cart.OperationType,
				&cart.CartID.OperationTime,
				&item.ProductID,
				&item.Quantity,
				&item.Price,
			)
		}
	}
	if err := r.session.ExecuteBatch(batch); err != nil {
		if logErr := r.log.Log("error", fmt.Sprintf("Failed to insert carriage trip info: %v", err)); logErr != nil {
			return fmt.Errorf("failed to execute batch and log error: %v (log error: %v)", err, logErr)
		}
		return fmt.Errorf("failed to execute batch: %w", err)
	}
	return nil
}

// GetEmployeeCartsInTrip Gets all carts employee has sold during trip, returns array of Carts
func (r *SalesRepository) GetEmployeeCartsInTrip(ctx context.Context, tripID *models.TripID, employeeID *string) ([]models.Cart, error) {
	iter := r.session.Query(getEmployeeCartsInTripQuery,
		&tripID.RouteID,
		&tripID.StartTime,
		&employeeID).Iter()

	carts, err := aggregateCartsFromRows(iter, *employeeID)
	if err != nil {
		_ = r.log.Log("error", fmt.Sprintf("Failed to aggregate carts by employee in trip %v", err))
		return nil, err
	}

	err = iter.Close()
	if err != nil {
		_ = r.log.Log("error", fmt.Sprintf("Failed to get carts by employee ID in trip %v", err))
		return nil, err
	}

	return carts, nil
}

// GetEmployeeIDsByTrip Gets all employees by TripID (RouteID, StartTime)
func (r *SalesRepository) GetEmployeeIDsByTrip(ctx context.Context, tripID *models.TripID) ([]string, error) {
	iter := r.session.Query(getEmployeeIdsByTripQuery,
		&tripID.RouteID,
		&tripID.StartTime).Iter()

	// Making a map to get unique ID's since Cassandra only allows DISTINCT for partition keys
	uniqueIDs := make(map[string]struct{})
	var employeeID string
	for iter.Scan(&employeeID) {
		uniqueIDs[employeeID] = struct{}{}
	}

	if err := iter.Close(); err != nil {
		_ = r.log.Log("error", fmt.Sprintf("Failed to get all employees by trip ID %v", err))
		return nil, err
	}

	var employeeIDs []string
	for id := range uniqueIDs {
		employeeIDs = append(employeeIDs, id)
	}

	return employeeIDs, nil
}

// UpdateItemQuantity Updates quantity of items in cart
func (r *SalesRepository) UpdateItemQuantity(ctx context.Context, tripID *models.TripID, cartID *models.CartID, productID *int, newQuantity *int16) error {
	applied, err := r.session.Query(updateItemQuantityQuery, newQuantity,
		tripID.RouteID,
		tripID.StartTime,
		cartID.EmployeeID,
		cartID.OperationTime,
		productID).ScanCAS()

	if err != nil {
		_ = r.log.Log("error", fmt.Sprintf("Failed to update item quantity %v", err))
		return err
	}
	if !applied {
		return fmt.Errorf("Transaction does not exist")
	}

	return nil
}

// DeleteItemFromCart Deletes cart item (operation)
func (r *SalesRepository) DeleteItemFromCart(tripID *models.TripID, cartID *models.CartID, productID *int) error {
	result := r.session.Query(deleteItemFromCartQuery,
		tripID.RouteID,
		tripID.StartTime,
		cartID.EmployeeID,
		cartID.OperationTime,
		productID).Exec()

	if result != nil {
		_ = r.log.Log("error", fmt.Sprintf("Failed to delete item in cart %v", result))
		return result
	}
	return nil
}

// Helper function to process rows and return an array of Carts
func aggregateCartsFromRows(iter *gocql.Iter, employeeID string) ([]models.Cart, error) {
	cartMap := make(map[string]*models.Cart)

	var operationTime time.Time
	var operationType int8
	var productID int
	var quantity int16
	var price int64

	for iter.Scan(&operationTime, &operationType, &productID, &quantity, &price) {

		// Create a cartID struct
		cartID := models.CartID{
			EmployeeID:    employeeID,
			OperationTime: operationTime,
		}

		//Declares key and ensures uniqueness by employeeID and operationTime
		cartKey := createCartKey(employeeID, operationTime)
		item := createCartItem(productID, quantity, price)

		// Checks if cart key exists in current map and inserts items into corresponding key
		if cart, exists := cartMap[cartKey]; exists {
			addItemToExistingCart(cart, item)
		} else { //Inserts new key and populates with items
			cartMap[cartKey] = createNewCart(cartID, operationType, item)
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

func createCartKey(employeeID string, operationTime time.Time) string {
	return fmt.Sprintf("%s-%s", employeeID, operationTime.Format(time.RFC3339))
}

func createCartItem(productID int, quantity int16, price int64) models.Item {
	return models.Item{
		ProductID: productID,
		Quantity:  quantity,
		Price:     price,
	}
}

func addItemToExistingCart(cart *models.Cart, item models.Item) {
	cart.Items = append(cart.Items, item)
}

func createNewCart(cartID models.CartID, operationType int8, item models.Item) *models.Cart {
	return &models.Cart{
		CartID:        cartID,
		OperationType: operationType,
		Items:         []models.Item{item},
	}
}
