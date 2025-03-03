package service

import (
	"ChaikaReports/internal/models"
	"ChaikaReports/internal/repository"
	"context"
)

type SalesService interface {
	InsertData(ctx context.Context, carriageReport *models.Carriage) error
	GetEmployeeCartsInTrip(ctx context.Context, tripID *models.TripID, employeeID *string) ([]models.Cart, error)
	GetEmployeeIDsByTrip(ctx context.Context, tripID *models.TripID) ([]string, error)
	UpdateItemQuantity(ctx context.Context, tripID *models.TripID, cartID *models.CartID, productID *int, newQuantity *int16) error
	DeleteItemFromCart(ctx context.Context, tripID *models.TripID, cartID *models.CartID, productID *int) error
}

type salesService struct {
	repo repository.SalesRepository
}

// NewSalesService Creates new salesService
func NewSalesService(repo repository.SalesRepository) SalesService {
	return &salesService{repo: repo}
}

// InsertData Inserts incoming carriageReport data
func (s *salesService) InsertData(ctx context.Context, carriageReport *models.Carriage) error {
	return s.repo.InsertData(ctx, carriageReport)
}

// GetEmployeeCartsInTrip Gets all carts an employee made during trip
func (s *salesService) GetEmployeeCartsInTrip(ctx context.Context, tripID *models.TripID, employeeID *string) ([]models.Cart, error) {
	return s.repo.GetEmployeeCartsInTrip(ctx, tripID, employeeID)
}

// GetEmployeeIDsByTrip Gets all employee ID's in trip
func (s *salesService) GetEmployeeIDsByTrip(ctx context.Context, tripID *models.TripID) ([]string, error) {
	return s.repo.GetEmployeeIDsByTrip(ctx, tripID)
}

// UpdateItemQuantity Updates item quantity in cart
func (s *salesService) UpdateItemQuantity(ctx context.Context, tripID *models.TripID, cartID *models.CartID, productID *int, newQuantity *int16) error {
	return s.repo.UpdateItemQuantity(ctx, tripID, cartID, productID, newQuantity)
}

// DeleteItemFromCart Deletes item from cart
func (s *salesService) DeleteItemFromCart(ctx context.Context, tripID *models.TripID, cartID *models.CartID, productID *int) error {
	return s.repo.DeleteItemFromCart(ctx, tripID, cartID, productID)
}
