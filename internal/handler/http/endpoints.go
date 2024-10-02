package http

import (
	"ChaikaReports/internal/models"
	"ChaikaReports/internal/service"
	"context"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/log"
	"net/http"
)

type Endpoints struct {
	InsertData endpoint.Endpoint
}

func MakeEndpoints(logger log.Logger, service service.SalesService) Endpoints {
	return Endpoints{
		InsertData: makeInsertDataEndpoint(logger, service),
	}
}

func makeInsertDataEndpoint(logger log.Logger, s service.SalesService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(models.Carriage)
		err := s.InsertData(ctx, &req)
		if err != nil {
			_ = logger.Log("error", fmt.Sprintf("Failed to make insert data endpoint", err))
			return nil, err
		}
		return http.StatusOK, nil
	}
}
