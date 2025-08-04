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

const (
	invalidStartTimeErrorMessage     = "invalid start_time format; must be RFC3339"
	invalidOperationTimeErrorMessage = "invalid operation_time format; must be RFC3339"
	invalidRequestTypeErrorMessage   = "invalid request type"
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
// @Router       /sale [post]
func MakeInsertSalesEndpoint(svc service.SalesService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		carriage, ok := request.(*models.CarriageReport)
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

// MakeGetEmployeeCartsInTripEndpoint handles getting carts for an employee in a trip
//
// @Summary      Get Employee Carts in Trip
// @Description  Returns all carts handled by a specific employee during a specific trip.
// @Tags         Sales
// @Accept       json
// @Produce      json
// @Param        route_id     query     string  true  "Route ID"
// @Param        year         query     string  true  "Year"
// @Param        start_time   query     string  true  "Trip Start Time in RFC3339 format"
// @Param        employee_id  query     string  true  "Employee ID"
// @Success      200          {object}  schemas.GetEmployeeCartsInTripResponse
// @Failure      400          {object}  schemas.ErrorResponse
// @Failure      500          {object}  schemas.ErrorResponse
// @Router       /trip/cart/employee [get]
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

// MakeGetEmployeeIDsByTripEndpoint handles getting employee IDs by trip
//
// @Summary      Get Employee IDs by Trip
// @Description  Returns all employee IDs who worked during a specific trip.
// @Tags         Sales
// @Accept       json
// @Produce      json
// @Param        route_id    query     string  true  "Route ID"
// @Param        year        query     string  true  "Year"
// @Param        start_time  query     string  true  "Trip Start Time in RFC3339 format"
// @Success      200         {object}  schemas.GetEmployeeIDsByTripResponse
// @Failure      400         {object}  schemas.ErrorResponse
// @Failure      500         {object}  schemas.ErrorResponse
// @Router       /trip/employee_id [get]
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

// MakeGetEmployeeTripsEndpoint handles getting trips by employee
//
// @Summary      Get Employee Trips
// @Description  Returns all trips completed by an employee during a year.
// @Tags         Sales
// @Accept       json
// @Produce      json
// @Param        employee_id  query     string  true  "Employee ID"
// @Param        year         query     string  true  "Year"
// @Success      200          {object}  schemas.GetEmployeeTripsResponse
// @Failure      400          {object}  schemas.ErrorResponse
// @Failure      500          {object}  schemas.ErrorResponse
// @Router       /trip/employee_trip [get]
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

// MakeUpdateItemQuantityEndpoint handles updating quantity of an item in a cart
//
// @Summary      Update Item Quantity
// @Description  Updates the quantity of a specific product in a cart.
// @Tags         Sales
// @Accept       json
// @Produce      json
// @Param        request  body      schemas.UpdateItemQuantityRequest  true  "Update Item Quantity Request"
// @Success      200      {object}  schemas.UpdateItemQuantityResponse
// @Failure      400      {object}  schemas.ErrorResponse
// @Failure      500      {object}  schemas.ErrorResponse
// @Router       /trip/cart/item/quantity [put]
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

// MakeDeleteItemFromCartEndpoint handles deleting an item from a cart
//
// @Summary      Delete Item from Cart
// @Description  Deletes a product from a specific cart.
// @Tags         Sales
// @Accept       json
// @Produce      json
// @Param        request  body      schemas.DeleteItemFromCartRequest  true  "Delete Item from Cart Request"
// @Success      200      {object}  schemas.DeleteItemFromCartResponse
// @Failure      400      {object}  schemas.ErrorResponse
// @Failure      500      {object}  schemas.ErrorResponse
// @Router       /trip/cart/item [delete]
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
