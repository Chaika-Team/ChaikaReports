package decoder

import (
	"ChaikaReports/internal/handler/http/schemas"
	"ChaikaReports/internal/models"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"net/http"
	"strconv"
	"time"
)

const invalidRequestBodyErrorMessage = "invalid request body"

// DecodeInsertSalesRequest decodes the HTTP request into the domain model
func DecodeInsertSalesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req schemas.InsertSalesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.New(invalidRequestBodyErrorMessage)
	}

	// Convert schemas.InsertSalesRequest to models.CarriageReport
	carriageStartTime, err := time.Parse(time.RFC3339, req.TripID.StartTime)
	if err != nil {
		return nil, errors.New("invalid trip start_time format")
	}

	req.TripID.Year = strconv.Itoa(carriageStartTime.Year())

	// Validate the request
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return nil, errors.New("validation failed: " + err.Error())
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
			if item.Quantity == 0 {
				return nil, errors.New("invalid item quantity")
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

	carriage := &models.CarriageReport{
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

func DecodeGetEmployeeCartsInTripRequest(_ context.Context, r *http.Request) (interface{}, error) {
	query := r.URL.Query()
	routeID := query.Get("route_id")
	year := query.Get("year")
	startTime := query.Get("start_time")
	employeeID := query.Get("employee_id")

	if routeID == "" || year == "" || startTime == "" || employeeID == "" {
		return nil, errors.New("missing one or more required query parameters: route_id, year, start_time, employee_id")
	}

	req := schemas.GetEmployeeCartsInTripRequest{
		TripID: schemas.TripID{
			RouteID:   routeID,
			Year:      year,
			StartTime: startTime,
		},
		EmployeeID: employeeID,
	}
	return req, nil
}

func DecodeGetEmployeeIDsByTripRequest(_ context.Context, r *http.Request) (interface{}, error) {
	query := r.URL.Query()
	routeID := query.Get("route_id")
	year := query.Get("year")
	startTime := query.Get("start_time")

	if routeID == "" || year == "" || startTime == "" {
		return nil, errors.New("missing required query parameters: route_id, year or start_time")
	}

	req := schemas.GetEmployeeIDsByTripRequest{
		TripID: schemas.TripID{
			RouteID:   routeID,
			Year:      year,
			StartTime: startTime,
		},
	}
	return req, nil
}

func DecodeGetEmployeeTripsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	query := r.URL.Query()
	employeeID := query.Get("employee_id")
	year := query.Get("year")

	if employeeID == "" || year == "" {
		return nil, errors.New("missing required query parameters: employee_id or year")
	}

	req := schemas.GetEmployeeTripsRequest{
		EmployeeID: employeeID,
		Year:       year,
	}
	return req, nil
}

func DecodeUpdateItemQuantityRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req schemas.UpdateItemQuantityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.New(invalidRequestBodyErrorMessage)
	}
	return req, nil
}

func DecodeDeleteItemFromCartRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req schemas.DeleteItemFromCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.New(invalidRequestBodyErrorMessage)
	}
	return req, nil
}
