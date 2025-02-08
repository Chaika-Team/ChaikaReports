package db_test_test

import (
	"ChaikaReports/internal/config"
	"ChaikaReports/internal/models"
	"ChaikaReports/internal/repository/cassandra"
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
)

var (
	testSession *gocql.Session
	testRepo    *cassandra.SalesRepository
	ctx         = context.Background()
)

func TestMain(m *testing.M) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "../../config.yml"
	}
	cfg, _ := config.LoadConfig(configPath)

	// Connect to the test keyspace
	var err error
	testSession, err = cassandra.InitCassandra(log.NewNopLogger(), cfg.CassandraTest.Keyspace, cfg.CassandraTest.Hosts, cfg.CassandraTest.User, cfg.CassandraTest.Password, cfg.CassandraTest.Timeout, cfg.CassandraTest.RetryDelay, cfg.CassandraTest.RetryAttempts)
	if err != nil {
		panic("Failed to connect to test keyspace")
	}

	testRepo = cassandra.NewSalesRepository(testSession, log.NewNopLogger())

	code := m.Run()

	cassandra.CloseCassandra(testSession)

	// Exit with the code from m.Run
	os.Exit(code)
}

func TestInsert(t *testing.T) {
	// Use testRepo to ensure it's referenced
	assert.NotNil(t, testRepo, "testRepo should be initialized")
	tripStartTime := time.Date(2023, 1, 15, 10, 0, 1, 0, time.UTC)

	carriage := &models.Carriage{
		TripID: models.TripID{
			RouteID:   "route_test",
			StartTime: tripStartTime,
		},
		EndTime:    tripStartTime.Add(1 * time.Hour),
		CarriageID: 10,
		Carts: []models.Cart{
			// First employee's cart
			{
				CartID: models.CartID{
					EmployeeID:    "67890", // Employee ID is a string
					OperationTime: time.Date(2023, 1, 15, 12, 30, 0, 0, time.UTC),
				},
				OperationType: 1,
				Items: []models.Item{
					{ProductID: 1, Quantity: 10, Price: 100.0},
					{ProductID: 2, Quantity: 5, Price: 200.0},
					{ProductID: 3, Quantity: 2, Price: 250.0},
				},
			},
			// Second employee's cart
			{
				CartID: models.CartID{
					EmployeeID:    "67890",
					OperationTime: time.Date(2023, 1, 15, 12, 35, 0, 0, time.UTC),
				},
				OperationType: 2,
				Items: []models.Item{
					{ProductID: 4, Quantity: 7, Price: 150.0},
					{ProductID: 5, Quantity: 3, Price: 80.0},
				},
			},
			// Third employee's cart
			{
				CartID: models.CartID{
					EmployeeID:    "98765",
					OperationTime: time.Date(2023, 1, 15, 12, 40, 0, 0, time.UTC),
				},
				OperationType: 1,
				Items: []models.Item{
					{ProductID: 6, Quantity: 1, Price: 500.0},
					{ProductID: 7, Quantity: 2, Price: 120.0},
					{ProductID: 8, Quantity: 4, Price: 75.0},
					{ProductID: 9, Quantity: 1, Price: 85.0},
				},
			},
			// Fourth employee's cart with multiple operations
			{
				CartID: models.CartID{
					EmployeeID:    "98765",
					OperationTime: time.Date(2023, 1, 15, 12, 45, 0, 0, time.UTC),
				},
				OperationType: 3,
				Items: []models.Item{
					{ProductID: 9, Quantity: 5, Price: 60.0},
				},
			},
			{
				CartID: models.CartID{
					EmployeeID:    "98765",
					OperationTime: time.Date(2023, 1, 15, 12, 50, 0, 0, time.UTC),
				},
				OperationType: 2,
				Items: []models.Item{
					{ProductID: 10, Quantity: 2, Price: 300.0},
				},
			},
		},
	}

	result := testRepo.InsertData(ctx, carriage)
	assert.NoError(t, result, "Failed to insert data")

}

func TestGetEmployeeCartsInTrip(t *testing.T) {
	// Define constants for test data
	var (
		testRouteID    = "route_test"
		testEmployeeID = "98765"
		testStartTime  = time.Date(2023, 1, 15, 10, 0, 1, 0, time.UTC)
	)

	// Define trip ID and employee ID
	tripID := models.TripID{
		RouteID:   testRouteID,
		StartTime: testStartTime,
	}
	employeeID := testEmployeeID

	// Retrieve carts from the repository
	carts, err := testRepo.GetEmployeeCartsInTrip(&tripID, &employeeID)

	// Log the returned carts for debugging
	t.Logf("Returned carts: %+v", carts)

	// Assert no error occurred
	assert.NoError(t, err, "Failed to get cart data for employee")

	// Assert the number of carts returned
	expectedCartCount := 3
	assert.Equal(t, expectedCartCount, len(carts), "Expected %d carts for employee %s", expectedCartCount, employeeID)

	// Define the expected carts
	expectedCarts := []models.Cart{
		{
			CartID: models.CartID{
				EmployeeID:    "98765",
				OperationTime: time.Date(2023, 1, 15, 12, 40, 0, 0, time.UTC),
			},
			OperationType: 1,
			Items: []models.Item{
				{ProductID: 6, Quantity: 1, Price: 500.0},
				{ProductID: 7, Quantity: 2, Price: 120.0},
				{ProductID: 8, Quantity: 4, Price: 75.0},
				{ProductID: 9, Quantity: 1, Price: 85.0},
			},
		},
		{
			CartID: models.CartID{
				EmployeeID:    "98765",
				OperationTime: time.Date(2023, 1, 15, 12, 45, 0, 0, time.UTC),
			},
			OperationType: 3,
			Items: []models.Item{
				{ProductID: 9, Quantity: 5, Price: 60.0},
			},
		},
		{
			CartID: models.CartID{
				EmployeeID:    "98765",
				OperationTime: time.Date(2023, 1, 15, 12, 50, 0, 0, time.UTC),
			},
			OperationType: 2,
			Items: []models.Item{
				{ProductID: 10, Quantity: 2, Price: 300.0},
			},
		},
	}

	// Iterate over expectedCarts and verify each one exists in the returned carts
	for _, expectedCart := range expectedCarts {
		found := false
		for _, actualCart := range carts {
			if actualCart.CartID.EmployeeID == expectedCart.CartID.EmployeeID &&
				actualCart.CartID.OperationTime.Equal(expectedCart.CartID.OperationTime) &&
				actualCart.OperationType == expectedCart.OperationType {

				// Assert that the items match
				assert.ElementsMatch(t, expectedCart.Items, actualCart.Items, "Items should match for cart with OperationTime %s", expectedCart.CartID.OperationTime)

				found = true
				break
			}
		}
		// Assert that the expected cart was found
		assert.True(t, found, "Expected cart not found: %+v", expectedCart)
	}
}
