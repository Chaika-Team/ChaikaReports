package service

import (
	"ChaikaReports/internal/models"
	"ChaikaReports/internal/repository"
	"context"
)

type SalesService interface {
	InsertData(ctx context.Context, carriageReport *models.CarriageReport) error
	GetTrip(ctx context.Context, tripID *models.TripID) (models.Trip, error)
	GetEmployeeCartsInTrip(ctx context.Context, tripID *models.TripID, employeeID *string) ([]models.Cart, error)
	GetEmployeeIDsByTrip(ctx context.Context, tripID *models.TripID) ([]string, error)
	GetEmployeeTrips(ctx context.Context, employeeID string, year string) ([]models.EmployeeTrip, error)
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
func (s *salesService) InsertData(ctx context.Context, carriageReport *models.CarriageReport) error {
	return s.repo.InsertData(ctx, carriageReport)
}

// GetTrip Gets all reports from a single trip
func (s *salesService) GetTrip(ctx context.Context, tripID *models.TripID) (models.Trip, error) {
	return s.repo.GetTrip(ctx, tripID)
}

// GetEmployeeCartsInTrip Gets all carts an employee made during trip
func (s *salesService) GetEmployeeCartsInTrip(ctx context.Context, tripID *models.TripID, employeeID *string) ([]models.Cart, error) {
	return s.repo.GetEmployeeCartsInTrip(ctx, tripID, employeeID)
}

// GetEmployeeIDsByTrip Gets all employee ID's in trip
func (s *salesService) GetEmployeeIDsByTrip(ctx context.Context, tripID *models.TripID) ([]string, error) {
	return s.repo.GetEmployeeIDsByTrip(ctx, tripID)
}

func (s *salesService) GetEmployeeTrips(ctx context.Context, employeeID string, year string) ([]models.EmployeeTrip, error) {
	return s.repo.GetEmployeeTrips(ctx, employeeID, year)
}

// UpdateItemQuantity Updates item quantity in cart
func (s *salesService) UpdateItemQuantity(ctx context.Context, tripID *models.TripID, cartID *models.CartID, productID *int, newQuantity *int16) error {
	return s.repo.UpdateItemQuantity(ctx, tripID, cartID, productID, newQuantity)
}

// DeleteItemFromCart Deletes item from cart
func (s *salesService) DeleteItemFromCart(ctx context.Context, tripID *models.TripID, cartID *models.CartID, productID *int) error {
	return s.repo.DeleteItemFromCart(ctx, tripID, cartID, productID)
}
