package http

import (
	"ChaikaReports/internal/handler/http/schemas"
	"ChaikaReports/internal/models"
	"ChaikaReports/internal/service"
	"context"
	"errors"
	"github.com/go-kit/kit/endpoint"
)

// MakeInsertSalesEndpoint creates the insert sales endpoint
func MakeInsertSalesEndpoint(svc service.SalesService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		carriage, ok := request.(*models.Carriage)
		if !ok {
			return nil, errors.New("invalid request type")
		}

		if err := svc.InsertData(ctx, carriage); err != nil {
			return nil, err
		}

		return schemas.InsertSalesResponse{
			Message: "Data inserted successfully",
		}, nil
	}
}
