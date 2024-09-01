package cassandra

import (
	"ChaikaReports/internal/models"
	"ChaikaReports/internal/repository"
)

type SalesRepository struct{}

func NewSalesRepository() repository.SalesRepository {
	return &SalesRepository{}
}

// Inserts all data from a CarriageReport into the Cassandra database
func (r *SalesRepository) InsertData(carriageReport *models.CarriageReport) error {
	for _, cart := range carriageReport.Carts {
		for _, item := range cart.Items {
			err := Session.Query(`
				INSERT INTO transactions (route_id, start_time, end_time, carriage_num, employee_id, operation_type, operation_time, product_id, quantity, price)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				&carriageReport.RouteID,
				&carriageReport.StartTime,
				&carriageReport.EndTime,
				&carriageReport.CarriageNum,
				&cart.EmployeeID,
				&cart.OperationType,
				&cart.OperationTime,
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

func (r *SalesRepository) GetEmployeeCarts() models.Operation {

}

func (r *SalesRepository) GetEmployeesByTrip() ([]models.SalesData, error) {

}

func (r *SalesRepository) UpdateItemQuantity() error {

}

func (r *SalesRepository) DeleteItemFromCart() error {

}
