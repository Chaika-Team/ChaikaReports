package http

import (
	"ChaikaReports/internal/handler/http/schemas"
	"ChaikaReports/internal/models"
	"ChaikaReports/internal/service"
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
)

// MakeInsertSalesEndpoint creates the insert sales endpoint.
//
// @Summary      Insert Sales Data
// @Description  Inserts sales data into the system.
// @Tags         Sales
// @Accept       json
// @Produce      json
// @Param        request  body      schemas.InsertSalesRequest  true  "Insert Sales Request"
// @Success      200      {object}  schemas.InsertSalesResponse "Data inserted successfully"
// @Failure      400      {object}  schemas.ErrorResponse       "Bad request"
// @Failure      500      {object}  schemas.ErrorResponse       "Internal server error"
// @Router       /sales [post]

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
