package http

import (
	"ChaikaReports/internal/handler/http/schemas"
	"ChaikaReports/internal/models"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
)

// DecodeInsertSalesRequest decodes the HTTP request into the domain model
func DecodeInsertSalesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req schemas.InsertSalesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.New("invalid request body")
	}

	// Validate the request
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return nil, errors.New("validation failed: " + err.Error())
	}

	// Convert schemas.InsertSalesRequest to models.Carriage
	carriageStartTime, err := time.Parse(time.RFC3339, req.TripID.StartTime)
	if err != nil {
		return nil, errors.New("invalid trip start_time format")
	}

	carriageEndTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return nil, errors.New("invalid end_time format")
	}

	var carts []models.Cart
	for _, cartSchema := range req.Carts {
		operationTime, err := time.Parse(time.RFC3339, cartSchema.CartID.OperationTime)
		if err != nil {
			return nil, errors.New("invalid cart operation_time format")
		}

		var items []models.Item
		for _, itemSchema := range cartSchema.Items {
			item := models.Item{
				ProductID: itemSchema.ProductID,
				Quantity:  itemSchema.Quantity,
				Price:     itemSchema.Price,
			}
			items = append(items, item)
		}

		cart := models.Cart{
			CartID: models.CartID{
				EmployeeID:    cartSchema.CartID.EmployeeID,
				OperationTime: operationTime,
			},
			OperationType: cartSchema.OperationType,
			Items:         items,
		}
		carts = append(carts, cart)
	}

	carriage := &models.Carriage{
		TripID: models.TripID{
			RouteID:   req.TripID.RouteID,
			StartTime: carriageStartTime,
		},
		EndTime:    carriageEndTime,
		CarriageID: req.CarriageID,
		Carts:      carts,
	}

	return carriage, nil
}
