package service

import (
	"ChaikaReports/internal/models"
	"ChaikaReports/internal/repository"
)

type SalesService interface {
	InsertData(carriageReport *models.Carriage) error
	GetEmployeeCartsInTrip(tripID *models.TripID, employeeID *string) ([]models.Cart, error)
	GetEmployeeIDsByTrip(tripID *models.TripID) ([]string, error)
	UpdateItemQuantity(tripID *models.TripID, cartID *models.CartID, productID *int, newQuantity *int16) error
	DeleteItemFromCart(tripID *models.TripID, cartID *models.CartID, productID *int) error
}

type salesService struct {
	repo repository.SalesRepository
}

// Creates new salesService
func NewSalesService(repo repository.SalesRepository) SalesService {
	return &salesService{repo: repo}
}

func (s *salesService) InsertData(carriageReport *models.Carriage) error {
	return s.repo.InsertData(carriageReport)
}

func (s *salesService) GetEmployeeCartsInTrip(tripID *models.TripID, employeeID *string) ([]models.Cart, error) {
	return s.repo.GetEmployeeCartsInTrip(tripID, employeeID)
}

func (s *salesService) GetEmployeeIDsByTrip(tripID *models.TripID) ([]string, error) {
	return s.GetEmployeeIDsByTrip(tripID)
}

func (s *salesService) UpdateItemQuantity(tripID *models.TripID, cartID *models.CartID, productID *int, newQuantity *int16) error {
	return s.UpdateItemQuantity(tripID, cartID, productID, newQuantity)
}

func (s *salesService) DeleteItemFromCart(tripID *models.TripID, cartID *models.CartID, productID *int) error {
	return s.DeleteItemFromCart(tripID, cartID, productID)
}
