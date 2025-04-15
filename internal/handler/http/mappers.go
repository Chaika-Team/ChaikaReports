package http

import (
	"ChaikaReports/internal/handler/http/schemas"
	"ChaikaReports/internal/models"
	"time"
)

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
