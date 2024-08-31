package cassandra

import (
	"ChaikaReports/internal/models"
	"ChaikaReports/internal/repository"
)

type SalesRepository struct{}

func NewSalesRepository() repository.SalesRepository {
	return &SalesRepository{}
}

func (r *SalesRepository) InsertData(operations models.Operations) error {

}

func (r *SalesRepository) GetEmployeeCarts() ([]models.Action, error) {

}

func (r *SalesRepository) GetEmployeesByTrip() ([]models.SalesData, error) {

}

func (r *SalesRepository) UpdateItemQuantity() error {

}

func (r *SalesRepository) DeleteItemFromCart() error {

}
