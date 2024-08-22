package service

import (
	"ChaikaReports/internal/models"
	"ChaikaReports/internal/repository"
)

type SalesService interface {
	InsertData(salesData *models.SalesData) error
	GetActionsByConductor(salesData *models.SalesData) ([]models.Action, error)
	GetConductorsByTripID(salesData *models.SalesData) ([]models.SalesData, error)
	UpdateActionCount(salesData *models.SalesData, action *models.Action) error
	DeleteProductFromAction(salesData *models.SalesData, action *models.Action) error
}

type salesService struct {
	repo repository.SalesRepository
}

func NewSalesService(repo repository.SalesRepository) SalesService {
	return &salesService{repo: repo}
}

func (s *salesService) InsertData(salesData *models.SalesData) error {
	return s.repo.InsertData(salesData)
}

func (s *salesService) GetActionsByConductor(salesData *models.SalesData) ([]models.Action, error) {
	return s.repo.GetActionsByConductor(salesData)
}

func (s *salesService) GetConductorsByTripID(salesData *models.SalesData) ([]models.SalesData, error) {
	return s.repo.GetConductorsByTripID(salesData)
}

func (s *salesService) UpdateActionCount(salesData *models.SalesData, action *models.Action) error {
	return s.repo.UpdateActionCount(salesData, action)
}

func (s *salesService) DeleteProductFromAction(salesData *models.SalesData, action *models.Action) error {
	return s.repo.DeleteProductFromAction(salesData, action)
}
