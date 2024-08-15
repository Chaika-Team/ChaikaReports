package cassandra

import (
	"ChaikaReports/internal/models"
	"ChaikaReports/internal/repository"
	"github.com/gocql/gocql"
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

func (r *SalesRepository) GetActionsByConductor(routeID, tripID, conductorID int) ([]models.Action, error) {
	var actions []models.Action
	iter := Session.Query(`SELECT product_id, operation_type_id, count, operation_time, operation_amount, operation_id FROM actions WHERE route_id = ? AND trip_id = ? AND conductor_id = ?`,
		routeID, tripID, conductorID).Iter()

	var action models.Action
	for iter.Scan(&action.ProductID, &action.OperationTypeID, &action.Count, &action.OperationTime, &action.OperationAmount, &action.OperationID) {
		actions = append(actions, action)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return actions, nil
}

func (r *SalesRepository) GetConductorsByTripID(routeID, tripID int) ([]models.SalesData, error) {
	var conductors []models.SalesData
	iter := Session.Query(`SELECT conductor_id FROM actions WHERE route_id = ? AND trip_id = ?`,
		routeID, tripID).Iter()

	var conductor models.SalesData
	for iter.Scan(&conductor.ConductorID) {
		conductors = append(conductors, conductor)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return conductors, nil
}

func (r *SalesRepository) UpdateActionCount(operationID gocql.UUID, routeID, tripID, productID, newCount int) error {
	return Session.Query(`UPDATE actions SET count = ? WHERE route_id = ? AND trip_id = ? AND operation_id = ? AND product_id = ?`,
		newCount, routeID, tripID, operationID, productID).Exec()
}

func (r *SalesRepository) DeleteProductFromAction(operationID gocql.UUID, routeID, tripID, conductorID, productID int) error {
	return Session.Query(`DELETE FROM actions WHERE route_id = ? AND trip_id = ? AND conductor_id = ? AND operation_id = ? AND product_id = ?`, routeID, tripID, conductorID, operationID, productID).Exec()
}
