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

func MakeGetEmployeeCartsInTripEndpoint(svc service.SalesService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// Assert the request to our schema type.
		req, ok := request.(schemas.GetEmployeeCartsInTripRequest)
		if !ok {
			return nil, errors.New("invalid request type")
		}

		// Parse the trip's start time.
		startTime, err := time.Parse(time.RFC3339, req.TripID.StartTime)
		if err != nil {
			return nil, errors.New("invalid start_time format; must be RFC3339")
		}

		// Build the domain TripID.
		tripID := models.TripID{
			RouteID:   req.TripID.RouteID,
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
			return nil, errors.New("invalid request type")
		}

		// Parse the trip's start time.
		startTime, err := time.Parse(time.RFC3339, req.TripID.StartTime)
		if err != nil {
			return nil, errors.New("invalid start_time format; must be RFC3339")
		}

		// Build the domain TripID.
		tripID := models.TripID{
			RouteID:   req.TripID.RouteID,
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

func MakeUpdateItemQuantityEndpoint(svc service.SalesService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(schemas.UpdateItemQuantityRequest)
		if !ok {
			return nil, errors.New("invalid request type")
		}

		// Parse TripID StartTime from string to time.Time.
		startTime, err := time.Parse(time.RFC3339, req.TripID.StartTime)
		if err != nil {
			return nil, errors.New("invalid start_time format; must be RFC3339")
		}

		tripID := models.TripID{
			RouteID:   req.TripID.RouteID,
			StartTime: startTime,
		}

		// Parse CartID OperationTime from string to time.Time.
		operationTime, err := time.Parse(time.RFC3339, req.CartID.OperationTime)
		if err != nil {
			return nil, errors.New("invalid operation_time format; must be RFC3339")
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
			return nil, errors.New("invalid request type")
		}

		// Parse Trip start time.
		startTime, err := time.Parse(time.RFC3339, req.TripID.StartTime)
		if err != nil {
			return nil, errors.New("invalid start_time format; must be RFC3339")
		}
		tripID := models.TripID{
			RouteID:   req.TripID.RouteID,
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

/* - - - - MAPPERS - - - - */

// mapDomainCartToSchemaCart converts a domain Cart (models.Cart) into a schema Cart (schemas.Cart).
func mapDomainCartToSchemaCart(cart models.Cart) schemas.Cart {
	return schemas.Cart{
		CartID: schemas.CartID{
			EmployeeID:    cart.CartID.EmployeeID,
			OperationTime: cart.CartID.OperationTime.Format(time.RFC3339), // Assuming domain CartID.OperationTime is time.Time.
		},
		OperationType: cart.OperationType,
		Items:         mapDomainItemsToSchemaItems(cart.Items),
	}
}

// mapDomainItemsToSchemaItems converts a slice of domain Items into schema Items.
func mapDomainItemsToSchemaItems(items []models.Item) []schemas.Item {
	var schemaItems []schemas.Item
	for _, item := range items {
		schemaItems = append(schemaItems, schemas.Item{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}
	return schemaItems
}
