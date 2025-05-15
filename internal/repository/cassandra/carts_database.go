package cassandra

import (
	"ChaikaReports/internal/models"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-kit/log"
	"github.com/gocql/gocql"
)

type SalesRepository struct {
	session CassandraSession
	log     log.Logger
}

func NewSalesRepository(session CassandraSession, logger log.Logger) *SalesRepository {
	return &SalesRepository{
		session: session,
		log:     logger,
	}
}

const insertOperationQuery = `
	INSERT INTO operations (
		route_id,
	    year,
	    start_time,
		end_time,
	    carriage_id, 
		employee_id,
		operation_type,
	    operation_time,
		product_id,
	    quantity,
		price)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

const insertEmployeeTripsQuery = `
	INSERT INTO employee_trips (
    	employee_id,
    	year,
		route_id,
		start_time,
		end_time)
	VALUES (?, ?, ?, ?, ?)`

const insertCompletedTripQuery = `
	INSERT INTO completed_trips (
	    route_id,
	    start_time,
	    year,
	    synchronized)
	VALUES (?,?,?,?)`

const insertRouteQuery = `
	INSERT INTO routes (
	    route_id)
	VALUES (?)`

const getTripQuery = `SELECT route_id, start_time, employee_id, operation_time, product_id, carriage_id, end_time, operation_type, price, quantity
	FROM operations
	WHERE route_id = ?
	AND year = ?
	AND start_time = ?
`

const getEmployeeCartsInTripQuery = `SELECT operation_time, operation_type, product_id, quantity, price
	FROM operations
	WHERE route_id = ?
	  AND year = ?
	  AND start_time = ?
	  AND employee_id = ?`

const getEmployeeIdsByTripQuery = `SELECT employee_id 
	FROM operations 
	WHERE route_id = ?
	  AND year = ?
	  AND start_time = ?`

const getEmployeeTripsQuery = `SELECT route_id, start_time, end_time
	FROM employee_trips
	WHERE employee_id = ?
      AND year = ?`

const updateItemQuantityQuery = `UPDATE operations SET quantity = ? 
    WHERE route_id = ? 
      AND year = ?
      AND start_time = ?
      AND employee_id = ?
      AND operation_time = ?
      AND product_id = ?
    IF EXISTS`

const deleteItemFromCartQuery = `DELETE FROM operations WHERE route_id = ? AND year = ? AND start_time = ? AND employee_id = ? AND operation_time = ? AND product_id = ? IF EXISTS`

// InsertData Inserts all data from a CarriageReport into the Cassandra database
func (r *SalesRepository) InsertData(ctx context.Context, carriageReport *models.CarriageReport) error {
	batch := r.session.NewBatch(gocql.LoggedBatch).WithContext(ctx)
	carriageReport.TripID.Year = strconv.Itoa(carriageReport.TripID.StartTime.Year())
	for _, cart := range carriageReport.Carts {
		batch.Query(insertEmployeeTripsQuery,
			&cart.CartID.EmployeeID,
			&carriageReport.TripID.Year,
			&carriageReport.TripID.RouteID,
			&carriageReport.TripID.StartTime,
			&carriageReport.EndTime,
		)

		batch.Query(insertCompletedTripQuery,
			&carriageReport.TripID.RouteID,
			&carriageReport.TripID.StartTime,
			&carriageReport.TripID.Year,
			false,
		)

		batch.Query(insertRouteQuery,
			&carriageReport.TripID.RouteID,
		)

		for _, item := range cart.Items {
			// Batch query allows to save data integrity by stopping transaction if at least one insertion fails
			batch.Query(insertOperationQuery,
				&carriageReport.TripID.RouteID,
				&carriageReport.TripID.Year,
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

	err := r.session.ExecuteBatch(batch)
	if err != nil {
		_ = r.log.Log("error", fmt.Sprintf("Failed to insert carriage trip info: %v", err))
		return fmt.Errorf("failed to execute batch: %w", err)
	}
	return nil
}

func (r *SalesRepository) GetTrip(ctx context.Context, tripID *models.TripID) (models.Trip, error) {
	// Fire the query
	iter := r.session.
		Query(getTripQuery, tripID.RouteID, tripID.Year, tripID.StartTime).
		WithContext(ctx).
		Iter()

	// We'll build two maps: one from carriageID → *models.CarriageReport,
	// and one from carriageID → map[cartKey]*models.Cart
	carriageMap := make(map[int8]*models.CarriageReport, 0)
	cartMaps := make(map[int8]map[string]*models.Cart, 0)

	// Row‐scan variables
	var (
		_routeID   string
		_startTime time.Time
		empID      string
		opTime     time.Time
		prodID     int
		carriageID int8
		endTime    time.Time
		opType     int8
		price      int64
		quantity   int16
	)

	// Iterate through every operation in this trip
	for iter.Scan(
		&_routeID,
		&_startTime,
		&empID,
		&opTime,
		&prodID,
		&carriageID,
		&endTime,
		&opType,
		&price,
		&quantity,
	) {
		// 1) ensure we have a CarriageReport object
		car, ok := carriageMap[carriageID]
		if !ok {
			car = &models.CarriageReport{
				TripID:     *tripID,
				EndTime:    endTime,
				CarriageID: carriageID,
			}
			carriageMap[carriageID] = car
			cartMaps[carriageID] = make(map[string]*models.Cart)
		}

		// 2) aggregate into the right cart
		cm := cartMaps[carriageID]
		key := createCartKey(empID, opTime)

		item := createCartItem(prodID, quantity, price)
		if existingCart, found := cm[key]; found {
			addItemToExistingCart(existingCart, item)
		} else {
			cartID := models.CartID{
				EmployeeID:    empID,
				OperationTime: opTime,
			}
			cm[key] = createNewCart(cartID, opType, item)
		}
	}

	// Close the iterator and catch any error
	if err := iter.Close(); err != nil {
		_ = r.log.Log("error", fmt.Sprintf("GetTrip: iter.Close failed: %v", err))
		return models.Trip{}, err
	}

	// 3) stitch carts back into each CarriageReport, then collect into the Trip
	var trip models.Trip
	for cid, car := range carriageMap {
		cm := cartMaps[cid]
		carts := make([]models.Cart, 0, len(cm))
		for _, c := range cm {
			carts = append(carts, *c)
		}
		car.Carts = carts
		trip.Carriage = append(trip.Carriage, *car)
	}

	return trip, nil
}

// GetEmployeeCartsInTrip Gets all carts employee has sold during trip, returns array of Carts
func (r *SalesRepository) GetEmployeeCartsInTrip(ctx context.Context, tripID *models.TripID, employeeID *string) ([]models.Cart, error) {
	iter := r.session.Query(getEmployeeCartsInTripQuery,
		&tripID.RouteID,
		&tripID.Year,
		&tripID.StartTime,
		&employeeID).WithContext(ctx).Iter()

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
		&tripID.Year,
		&tripID.StartTime).WithContext(ctx).Iter()

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

func (r *SalesRepository) GetEmployeeTrips(ctx context.Context, employeeID string, year string) ([]models.EmployeeTrip, error) {
	iter := r.session.Query(getEmployeeTripsQuery,
		employeeID,
		year).WithContext(ctx).Iter()

	var routeID string
	var startTime, endTime time.Time
	var employeeTrips []models.EmployeeTrip

	for iter.Scan(&routeID, &startTime, &endTime) {
		employeeTrip := models.EmployeeTrip{
			EmployeeID: employeeID,
			TripID: models.TripID{
				RouteID:   routeID,
				Year:      year,
				StartTime: startTime,
			},
			EndTime: endTime,
		}
		employeeTrips = append(employeeTrips, employeeTrip)
	}

	if err := iter.Close(); err != nil {
		_ = r.log.Log("error", fmt.Sprintf("Failed to get employee trips: %v", err))
		return nil, err
	}

	return employeeTrips, nil
}

// UpdateItemQuantity Updates quantity of items in cart
func (r *SalesRepository) UpdateItemQuantity(ctx context.Context, tripID *models.TripID, cartID *models.CartID, productID *int, newQuantity *int16) error {
	applied, err := r.session.Query(updateItemQuantityQuery, newQuantity,
		tripID.RouteID,
		tripID.Year,
		tripID.StartTime,
		cartID.EmployeeID,
		cartID.OperationTime,
		productID).WithContext(ctx).ScanCAS()

	if err != nil {
		_ = r.log.Log("error", fmt.Sprintf("Failed to update item quantity %v", err))
		return err
	}
	if !applied {
		return fmt.Errorf("transaction does not exist")
	}

	return nil
}

// DeleteItemFromCart Deletes cart item (operation)
func (r *SalesRepository) DeleteItemFromCart(ctx context.Context, tripID *models.TripID, cartID *models.CartID, productID *int) error {
	deleted, err := r.session.Query(deleteItemFromCartQuery,
		tripID.RouteID,
		tripID.Year,
		tripID.StartTime,
		cartID.EmployeeID,
		cartID.OperationTime,
		productID).WithContext(ctx).ScanCAS()

	if err != nil {
		_ = r.log.Log("error", fmt.Sprintf("Failed to delete item in cart %v", err))
		return err
	}

	if !deleted {
		return fmt.Errorf("item does not exist")
	}

	return nil
}

// Helper function to process rows and return an array of Carts
func aggregateCartsFromRows(iter Iter, employeeID string) ([]models.Cart, error) {
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
