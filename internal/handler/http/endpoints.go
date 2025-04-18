package http

import (
	"ChaikaReports/internal/handler/http/schemas"
	"ChaikaReports/internal/models"
	"ChaikaReports/internal/service"
	"context"
	"errors"
	"github.com/go-kit/kit/endpoint"
	"time"
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

const (
	invalidStartTimeErrorMessage     = "invalid start_time format; must be RFC3339"
	invalidOperationTimeErrorMessage = "invalid operation_time format; must be RFC3339"
	invalidRequestTypeErrorMessage   = "invalid request type"
)

// MakeInsertSalesEndpoint creates the insert sales endpoint
func MakeInsertSalesEndpoint(svc service.SalesService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		carriage, ok := request.(*models.Carriage)
		if !ok {
			return nil, errors.New(invalidRequestTypeErrorMessage)
		}

		err := svc.InsertData(ctx, carriage)
		if err != nil {
			return nil, err
		}

		return schemas.InsertSalesResponse{
			Message: "Data inserted successfully",
		}, nil
	}
}

func MakeGetEmployeeCartsInTripEndpoint(svc service.SalesService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// Assert the request to our schema type.
		req, ok := request.(schemas.GetEmployeeCartsInTripRequest)
		if !ok {
			return nil, errors.New(invalidRequestTypeErrorMessage)
		}

		// Parse the trip's start time.
		startTime, err := time.Parse(time.RFC3339, req.TripID.StartTime)
		if err != nil {
			return nil, errors.New(invalidStartTimeErrorMessage)
		}

		// Build the domain TripID.
		tripID := models.TripID{
			RouteID:   req.TripID.RouteID,
			Year:      req.TripID.Year,
			StartTime: startTime,
		}

		// Call the service. EmployeeID remains a string.
		carts, err := svc.GetEmployeeCartsInTrip(ctx, &tripID, &req.EmployeeID)
		if err != nil {
			return nil, err
		}

		// Map domain carts to our schema response.
		var schemaCarts []schemas.Cart
		for _, cart := range carts {
			schemaCarts = append(schemaCarts, mapDomainCartToSchemaCart(cart))
		}

		response := schemas.GetEmployeeCartsInTripResponse{
			Carts: schemaCarts,
		}
		return response, nil
	}
}

func MakeGetEmployeeIDsByTripEndpoint(svc service.SalesService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(schemas.GetEmployeeIDsByTripRequest)
		if !ok {
			return nil, errors.New(invalidRequestTypeErrorMessage)
		}

		// Parse the trip's start time.
		startTime, err := time.Parse(time.RFC3339, req.TripID.StartTime)
		if err != nil {
			return nil, errors.New(invalidStartTimeErrorMessage)
		}

		// Build the domain TripID.
		tripID := models.TripID{
			RouteID:   req.TripID.RouteID,
			Year:      req.TripID.Year,
			StartTime: startTime,
		}

		// Call the service method.
		employeeIDs, err := svc.GetEmployeeIDsByTrip(ctx, &tripID)
		if err != nil {
			return nil, err
		}

		// Build and return the response.
		return schemas.GetEmployeeIDsByTripResponse{
			EmployeeIDs: employeeIDs,
		}, nil
	}
}

// MakeGetEmployeeTripsEndpoint creates the endpoint for fetching an employee's trips.
func MakeGetEmployeeTripsEndpoint(svc service.SalesService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// Assert the request type to our schema type.
		req, ok := request.(schemas.GetEmployeeTripsRequest)
		if !ok {
			return nil, errors.New("invalid request type")
		}

		// Call the service method to get the trips for the given employee and year.
		trips, err := svc.GetEmployeeTrips(ctx, req.EmployeeID, req.Year)
		if err != nil {
			return nil, err
		}

		// Map each domain EmployeeTrip to the schema EmployeeTrip.
		var schemaTrips []schemas.EmployeeTrip
		for _, t := range trips {
			// We need to convert the domain TripID.StartTime (assumed to be time.Time)
			// into a string in RFC3339 format, as defined by the schema.
			schemaTrip := schemas.EmployeeTrip{
				EmployeeID: t.EmployeeID,
				TripID: schemas.TripID{
					RouteID:   t.TripID.RouteID,
					Year:      t.TripID.Year,
					StartTime: t.TripID.StartTime.Format(time.RFC3339),
				},
				EndTime: t.EndTime, // EndTime remains as time.Time per your schema.
			}
			schemaTrips = append(schemaTrips, schemaTrip)
		}

		return schemas.GetEmployeeTripsResponse{
			EmployeeTrips: schemaTrips,
		}, nil
	}
}

func MakeUpdateItemQuantityEndpoint(svc service.SalesService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(schemas.UpdateItemQuantityRequest)
		if !ok {
			return nil, errors.New(invalidRequestTypeErrorMessage)
		}

		// Parse TripID StartTime from string to time.Time.
		startTime, err := time.Parse(time.RFC3339, req.TripID.StartTime)
		if err != nil {
			return nil, errors.New(invalidStartTimeErrorMessage)
		}

		tripID := models.TripID{
			RouteID:   req.TripID.RouteID,
			Year:      req.TripID.Year,
			StartTime: startTime,
		}

		// Parse CartID OperationTime from string to time.Time.
		operationTime, err := time.Parse(time.RFC3339, req.CartID.OperationTime)
		if err != nil {
			return nil, errors.New(invalidOperationTimeErrorMessage)
		}

		cartID := models.CartID{
			EmployeeID:    req.CartID.EmployeeID,
			OperationTime: operationTime,
		}

		// Call the service method.
		err = svc.UpdateItemQuantity(ctx, &tripID, &cartID, &req.ProductID, &req.NewQuantity)
		if err != nil {
			return nil, err
		}

		return schemas.UpdateItemQuantityResponse{
			Message: "Item quantity updated successfully",
		}, nil
	}
}

func MakeDeleteItemFromCartEndpoint(svc service.SalesService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(schemas.DeleteItemFromCartRequest)
		if !ok {
			return nil, errors.New(invalidRequestTypeErrorMessage)
		}

		// Parse Trip start time.
		startTime, err := time.Parse(time.RFC3339, req.TripID.StartTime)
		if err != nil {
			return nil, errors.New(invalidStartTimeErrorMessage)
		}
		tripID := models.TripID{
			RouteID:   req.TripID.RouteID,
			Year:      req.TripID.Year,
			StartTime: startTime,
		}

		// Parse Cart operation time.
		operationTime, err := time.Parse(time.RFC3339, req.CartID.OperationTime)
		if err != nil {
			return nil, errors.New("invalid operation_time format; must be RFC3339")
		}
		cartID := models.CartID{
			EmployeeID:    req.CartID.EmployeeID,
			OperationTime: operationTime,
		}

		// Call the service method.
		err = svc.DeleteItemFromCart(ctx, &tripID, &cartID, &req.ProductID)
		if err != nil {
			return nil, err
		}

		return schemas.DeleteItemFromCartResponse{
			Message: "Item deleted successfully",
		}, nil
	}
}

/* - - - - HELPER FUNCTIONS - - - - */
