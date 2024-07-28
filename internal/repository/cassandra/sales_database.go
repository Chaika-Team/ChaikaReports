package cassandra

import (
	"ChaikaReports/internal/models"
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
)

type SalesRepository struct {
	client Client
	log    log.Logger
}

// NewSalesRepository creates a new instance of SalesRepository.
func NewSalesRepository(client Client, logger log.Logger) *SalesRepository {
	return &SalesRepository{
		client: client,
		log:    logger,
	}
}

func (r *SalesRepository) InsertSalesData(ctx context.Context, salesData *models.SalesData) error {
	// Insert into sales_data table
	sql := `INSERT INTO sales_data (trip_id, carriage_id, conductor_id) VALUES (?, ?, ?)`
	err := r.client.Exec(ctx, sql, salesData.TripID, salesData.CarriageID, salesData.ConductorID)
	if err != nil {
		_ = r.log.Log("error", fmt.Sprintf("Failed to insert into sales_data: %v", err))
		return err
	}

	// Insert into actions table
	for _, action := range salesData.Actions {
		sql := `INSERT INTO actions (trip_id, carriage_id, conductor_id, product_id, operation_type_id, count) VALUES (?, ?, ?, ?, ?, ?)`
		err := r.client.Exec(ctx, sql, salesData.TripID, salesData.CarriageID, salesData.ConductorID, action.ProductID, action.OperationTypeID, action.Count)
		if err != nil {
			_ = r.log.Log("error", fmt.Sprintf("Failed to insert into actions: %v", err))
			return err
		}
	}
	return nil
}

func (r *SalesRepository) GetActionsByConductor(ctx context.Context, tripID, conductorID int) ([]models.Action, error) {
	sql := `SELECT product_id, operation_type_id, count FROM actions WHERE trip_id = ? AND conductor_id = ?`
	iter, err := r.client.Query(ctx, sql, tripID, conductorID)
	if err != nil {
		_ = r.log.Log("error", fmt.Sprintf("Failed to get actions by conductor: %v", err))
		return nil, err
	}

	var actions []models.Action
	var action models.Action
	for iter.Scan(&action.ProductID, &action.OperationTypeID, &action.Count) {
		actions = append(actions, action)
	}
	if err := iter.Close(); err != nil {
		_ = r.log.Log("error", fmt.Sprintf("Failed during rows iteration: %v", err))
		return nil, err
	}
	return actions, nil
}

func (r *SalesRepository) UpdateActionCount(ctx context.Context, tripID, carriageID, conductorID, productID, operationTypeID, newCount int) error {
	sql := `UPDATE actions SET count = ? WHERE trip_id = ? AND carriage_id = ? AND conductor_id = ? AND product_id = ? AND operation_type_id = ?`
	err := r.client.Exec(ctx, sql, newCount, tripID, carriageID, conductorID, productID, operationTypeID)
	if err != nil {
		_ = r.log.Log("error", fmt.Sprintf("Failed to update action count: %v", err))
	}
	return err
}

func (r *SalesRepository) DeleteActions(ctx context.Context, tripID, conductorID int) error {
	sql := `DELETE FROM actions WHERE trip_id = ? AND conductor_id = ?`
	err := r.client.Exec(ctx, sql, tripID, conductorID)
	if err != nil {
		_ = r.log.Log("error", fmt.Sprintf("Failed to delete actions: %v", err))
	}
	return err
}
