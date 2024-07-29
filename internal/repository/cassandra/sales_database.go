package cassandra

import (
	"ChaikaReports/internal/models"
	"ChaikaReports/internal/repository"
)

type SalesRepository struct{}

func NewSalesRepository() repository.SalesRepository {
	return &SalesRepository{}
}

func (r *SalesRepository) InsertSalesData(salesData *models.SalesData) error {
	// Insert into sales_data table
	if err := Session.Query(`INSERT INTO sales_data (trip_id, carriage_id, conductor_id) VALUES (?, ?, ?)`,
		salesData.TripID, salesData.CarriageID, salesData.ConductorID).Exec(); err != nil {
		return err
	}

	// Insert into actions table
	for _, action := range salesData.Actions {
		if err := Session.Query(`INSERT INTO actions (trip_id, carriage_id, conductor_id, product_id, operation_type_id, count) VALUES (?, ?, ?, ?, ?, ?)`,
			salesData.TripID, salesData.CarriageID, salesData.ConductorID, action.ProductID, action.OperationTypeID, action.Count).Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (r *SalesRepository) GetActionsByConductor(tripID, conductorID int) ([]models.Action, error) {
	var actions []models.Action
	iter := Session.Query(`SELECT product_id, operation_type_id, count FROM actions WHERE trip_id = ? AND conductor_id = ? ALLOW FILTERING`,
		tripID, conductorID).Iter()

	var action models.Action
	for iter.Scan(&action.ProductID, &action.OperationTypeID, &action.Count) {
		actions = append(actions, action)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return actions, nil
}

func (r *SalesRepository) GetConductorsByTripID(tripID int) ([]models.SalesData, error) {
	var salesDataList []models.SalesData
	iter := Session.Query(`SELECT trip_id, carriage_id, conductor_id FROM sales_data WHERE trip_id = ?`, tripID).Iter()

	var salesData models.SalesData
	for iter.Scan(&salesData.TripID, &salesData.CarriageID, &salesData.ConductorID) {
		// Retrieve actions for each conductor in the trip
		actions, err := r.GetActionsByConductor(tripID, salesData.ConductorID)
		if err != nil {
			return nil, err
		}
		salesData.Actions = actions
		salesDataList = append(salesDataList, salesData)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return salesDataList, nil
}

func (r *SalesRepository) UpdateActionCount(tripID, carriageID, conductorID, productID, operationTypeID, newCount int) error {
	return Session.Query(`UPDATE actions SET count = ? WHERE trip_id = ? AND carriage_id = ? AND conductor_id = ? AND product_id = ? AND operation_type_id = ?`,
		newCount, tripID, carriageID, conductorID, productID, operationTypeID).Exec()
}

func (r *SalesRepository) DeleteActions(tripID, conductorID int) error {
	return Session.Query(`DELETE FROM actions WHERE trip_id = ? AND conductor_id = ?`, tripID, conductorID).Exec()
}
