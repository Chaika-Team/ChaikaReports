package cassandra

import (
	"ChaikaReports/internal/models"
	"ChaikaReports/internal/repository"
)

type SalesRepository struct{}

func NewSalesRepository() repository.SalesRepository {
	return &SalesRepository{}
}

func (r *SalesRepository) InsertData(salesData *models.SalesData) error {
	for _, action := range salesData.Actions {
		if err := Session.Query(`INSERT INTO actions (route_id, operation_time, carriage_id, conductor_id, count, operation_amount, operation_id, operation_type_id, product_id, trip_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			&salesData.RouteID, &action.OperationTime, &salesData.CarriageID, &salesData.ConductorID, &action.Count, &action.OperationAmount, &action.OperationID, &action.OperationTypeID, &action.ProductID, &salesData.TripID).Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (r *SalesRepository) GetActionsByConductor(salesData *models.SalesData) ([]models.Action, error) {
	var actions []models.Action
	iter := Session.Query(`SELECT product_id, operation_type_id, count, operation_time, operation_amount, operation_id FROM actions WHERE route_id = ? AND trip_id = ? AND conductor_id = ?`,
		&salesData.RouteID, &salesData.TripID, &salesData.ConductorID).Iter()

	var action models.Action
	for iter.Scan(&action.ProductID, &action.OperationTypeID, &action.Count, &action.OperationTime, &action.OperationAmount, &action.OperationID) {
		actions = append(actions, action)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return actions, nil
}

func (r *SalesRepository) GetConductorsByTripID(salesData *models.SalesData) ([]models.SalesData, error) {
	var conductors []models.SalesData
	iter := Session.Query(`SELECT conductor_id FROM actions WHERE route_id = ? AND trip_id = ?`,
		&salesData.RouteID, &salesData.TripID).Iter()

	var conductor models.SalesData
	for iter.Scan(&conductor.ConductorID) {
		conductors = append(conductors, conductor)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return conductors, nil
}

func (r *SalesRepository) UpdateActionCount(salesData *models.SalesData, action *models.Action) error {
	return Session.Query(`UPDATE actions SET count = ? WHERE route_id = ? AND trip_id = ? AND operation_id = ? AND product_id = ?`,
		&action.Count, &salesData.RouteID, &salesData.TripID, &action.OperationID, &action.ProductID).Exec()
}

func (r *SalesRepository) DeleteProductFromAction(salesData *models.SalesData, action *models.Action) error {
	return Session.Query(`DELETE FROM actions WHERE route_id = ? AND trip_id = ? AND conductor_id = ? AND operation_id = ? AND product_id = ?`,
		&salesData.RouteID, &salesData.TripID, &salesData.ConductorID, &action.OperationID, &action.ProductID).Exec()
}
