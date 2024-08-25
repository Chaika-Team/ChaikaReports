package endpoints

import (
	"ChaikaReports/internal/models"
	"ChaikaReports/internal/service"
	"context"
	"github.com/go-kit/kit/endpoint"
)

// Endpoints struct that holds all the endpoints definitions.
type Endpoints struct {
	InsertDataEndpoint              endpoint.Endpoint
	GetActionsByConductorEndpoint   endpoint.Endpoint
	GetConductorsByTripIDEndpoint   endpoint.Endpoint
	UpdateActionCountEndpoint       endpoint.Endpoint
	DeleteProductFromActionEndpoint endpoint.Endpoint
}

// MakeEndpoints initializes all the endpoints for the SalesService.
func MakeEndpoints(svc service.SalesService) Endpoints {
	return Endpoints{
		InsertDataEndpoint:              MakeInsertDataEndpoint(svc),
		GetActionsByConductorEndpoint:   MakeGetActionsByConductorEndpoint(svc),
		GetConductorsByTripIDEndpoint:   MakeGetConductorsByTripIDEndpoint(svc),
		UpdateActionCountEndpoint:       MakeUpdateActionCountEndpoint(svc),
		DeleteProductFromActionEndpoint: MakeDeleteProductFromActionEndpoint(svc),
	}
}

// Request and Response Structs for each endpoints

type InsertDataRequest struct {
	SalesData *models.SalesData `json:"sales_data"`
}

type InsertDataResponse struct {
	Err error `json:"err,omitempty"`
}

type GetActionsByConductorRequest struct {
	SalesData *models.SalesData `json:"sales_data"`
}

type GetActionsByConductorResponse struct {
	Actions []models.Action `json:"actions,omitempty"`
	Err     error           `json:"err,omitempty"`
}

type GetConductorsByTripIDRequest struct {
	SalesData *models.SalesData `json:"sales_data"`
}

type GetConductorsByTripIDResponse struct {
	Conductors []models.SalesData `json:"conductors,omitempty"`
	Err        error              `json:"err,omitempty"`
}

type UpdateActionCountRequest struct {
	SalesData *models.SalesData `json:"sales_data"`
	Action    *models.Action    `json:"action"`
}

type UpdateActionCountResponse struct {
	Err error `json:"err,omitempty"`
}

type DeleteProductFromActionRequest struct {
	SalesData *models.SalesData `json:"sales_data"`
	Action    *models.Action    `json:"action"`
}

type DeleteProductFromActionResponse struct {
	Err error `json:"err,omitempty"`
}

// Endpoint implementations

func MakeInsertDataEndpoint(svc service.SalesService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(InsertDataRequest)
		err := svc.InsertData(req.SalesData)
		return InsertDataResponse{Err: err}, nil
	}
}

func MakeGetActionsByConductorEndpoint(svc service.SalesService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetActionsByConductorRequest)
		actions, err := svc.GetActionsByConductor(req.SalesData)
		return GetActionsByConductorResponse{Actions: actions, Err: err}, nil
	}
}

func MakeGetConductorsByTripIDEndpoint(svc service.SalesService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetConductorsByTripIDRequest)
		conductors, err := svc.GetConductorsByTripID(req.SalesData)
		return GetConductorsByTripIDResponse{Conductors: conductors, Err: err}, nil
	}
}

func MakeUpdateActionCountEndpoint(svc service.SalesService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UpdateActionCountRequest)
		err := svc.UpdateActionCount(req.SalesData, req.Action)
		return UpdateActionCountResponse{Err: err}, nil
	}
}

func MakeDeleteProductFromActionEndpoint(svc service.SalesService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DeleteProductFromActionRequest)
		err := svc.DeleteProductFromAction(req.SalesData, req.Action)
		return DeleteProductFromActionResponse{Err: err}, nil
	}
}
