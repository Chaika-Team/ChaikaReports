package cassandra_test

import (
	"ChaikaReports/internal/config"
	"ChaikaReports/internal/models"
	"ChaikaReports/internal/repository/cassandra"
	"context"
	"github.com/go-kit/log"
	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

var (
	testSession *gocql.Session
	testRepo    *cassandra.SalesRepository
	ctx         context.Context
)

func TestMain(m *testing.M) {
	cfg := config.LoadConfig("C:/Users/Greg/GolandProjects/ChaikaReports/config.yml")

	// Connect to the test keyspace
	var err error
	testSession, err = cassandra.InitCassandra(log.NewNopLogger(), cfg.CassandraTest.Keyspace, cfg.CassandraTest.Hosts, cfg.CassandraTest.User, cfg.CassandraTest.Password)
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
	tripID := models.TripID{
		RouteID:   "route-test",
		StartTime: time.Date(2023, 1, 15, 10, 0, 1, 0, time.UTC)}
	employeeID := "98765"
	carts, err := testRepo.GetEmployeeCartsInTrip(&tripID, &employeeID)
	assert.NoError(t, err, "Failed to get cart data for employee")
	assert.Equal(t, 3, len(carts), "Expected 3 carts for employee")
}
