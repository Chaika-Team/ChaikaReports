// repository_mock_test.go
package cassandra

import (
	"ChaikaReports/internal/models"
	"context"
	"fmt"
	"github.com/go-kit/log"
	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

// --- HELPER TYPES ---
// --- Used in TestGetEmployeeCartsInTrip ---
type fakeRow struct {
	operationTime time.Time
	operationType int8
	productID     int
	quantity      int16
	price         int64
}

type SimpleFakeIter struct {
	rows  []fakeRow
	index int
}

func (s *SimpleFakeIter) Scan(dest ...interface{}) bool {
	if s.index >= len(s.rows) {
		return false
	}
	row := s.rows[s.index]
	s.index++

	// Expect exactly 5 destinations.
	if len(dest) != 5 {
		return false
	}
	if ptr, ok := dest[0].(*time.Time); ok {
		*ptr = row.operationTime
	} else {
		return false
	}
	if ptr, ok := dest[1].(*int8); ok {
		*ptr = row.operationType
	} else {
		return false
	}
	if ptr, ok := dest[2].(*int); ok {
		*ptr = row.productID
	} else {
		return false
	}
	if ptr, ok := dest[3].(*int16); ok {
		*ptr = row.quantity
	} else {
		return false
	}
	if ptr, ok := dest[4].(*int64); ok {
		*ptr = row.price
	} else {
		return false
	}
	return true
}

// --- Used in TestGetEmployeeIDsByTrip ---

type simpleFakeIterString struct {
	employeeIDs []string
	index       int
}

func (s *simpleFakeIterString) Scan(dest ...interface{}) bool {
	if s.index >= len(s.employeeIDs) {
		return false
	}
	// We expect exactly one destination pointer.
	if len(dest) != 1 {
		return false
	}
	if ptr, ok := dest[0].(*string); ok {
		*ptr = s.employeeIDs[s.index]
	} else {
		return false
	}
	s.index++
	return true
}

func (s *simpleFakeIterString) Close() error {
	return nil
}

func (s *SimpleFakeIter) Close() error {
	return nil
}

// --- Used in TestGetEmployeeTrips ---

type simpleFakeIterTrip struct {
	rows []struct {
		routeID   string
		year      string
		startTime time.Time
		endTime   time.Time
	}
	index int
}

func (s *simpleFakeIterTrip) Scan(dest ...interface{}) bool {
	if s.index >= len(s.rows) {
		return false
	}
	row := s.rows[s.index]
	s.index++

	// Expect exactly 3 destination pointers.
	if len(dest) != 3 {
		return false
	}
	if ptr, ok := dest[0].(*string); ok {
		*ptr = row.routeID
	} else {
		return false
	}
	if ptr, ok := dest[1].(*time.Time); ok {
		*ptr = row.startTime
	} else {
		return false
	}
	if ptr, ok := dest[2].(*time.Time); ok {
		*ptr = row.endTime
	} else {
		return false
	}
	return true
}

func (s *simpleFakeIterTrip) Close() error {
	return nil
}

// --- Used in TestGetEmployeeCartsInTrip_AggregateCloseError ---

type fakeIterWithError struct {
	errorOnCall int
	callCount   int
}

func (f *fakeIterWithError) Scan(dest ...interface{}) bool {
	return false
}

func (f *fakeIterWithError) Close() error {
	f.callCount++
	if f.callCount == f.errorOnCall {
		return fmt.Errorf("iter close error")
	}
	return nil
}

// --- Used in TestGetEmployeeIDsByTrip_CloseError ---

type fakeIterStringWithCloseError struct {
	employeeIDs []string
	index       int
}

func (f *fakeIterStringWithCloseError) Scan(dest ...interface{}) bool {
	if f.index >= len(f.employeeIDs) {
		return false
	}
	if len(dest) != 1 {
		return false
	}
	if ptr, ok := dest[0].(*string); ok {
		*ptr = f.employeeIDs[f.index]
	} else {
		return false
	}
	f.index++
	return true
}

func (f *fakeIterStringWithCloseError) Close() error {
	return fmt.Errorf("close error")
}

// --- Used in TestGetEmployeeTrips_CloseError ---

type fakeIterTripWithCloseError struct{}

func (f *fakeIterTripWithCloseError) Scan(dest ...interface{}) bool {
	return false
}

func (f *fakeIterTripWithCloseError) Close() error {
	return fmt.Errorf("trip iter close error")
}

// --- MOCK TYPES ---

// MockSession implements cassandra.CassandraSession.
type MockSession struct {
	mock.Mock
}

func (m *MockSession) Query(stmt string, values ...interface{}) Query {
	args := m.Called(stmt, values)
	return args.Get(0).(Query)
}

func (m *MockSession) NewBatch(batchType gocql.BatchType) Batch {
	args := m.Called(batchType)
	return args.Get(0).(Batch)
}

func (m *MockSession) ExecuteBatch(batch Batch) error {
	args := m.Called(batch)
	return args.Error(0)
}

func (m *MockSession) Close() {
	m.Called()
}

// FakeQuery implements cassandra.Query.
type FakeQuery struct {
	mock.Mock
}

func (fq *FakeQuery) WithContext(ctx context.Context) Query {
	args := fq.Called(ctx)
	return args.Get(0).(Query)
}

func (fq *FakeQuery) Exec() error {
	args := fq.Called()
	return args.Error(0)
}

func (fq *FakeQuery) Iter() Iter {
	args := fq.Called()
	return args.Get(0).(Iter)
}

func (fq *FakeQuery) ScanCAS(dest ...interface{}) (bool, error) {
	args := fq.Called(dest)
	return args.Bool(0), args.Error(1)
}

// FakeIter implements cassandra.Iter.
type FakeIter struct {
	mock.Mock
}

func (fi *FakeIter) Scan(dest ...interface{}) bool {
	args := fi.Called(dest)
	return args.Bool(0)
}

func (fi *FakeIter) Close() error {
	args := fi.Called()
	return args.Error(0)
}

// FakeBatch implements cassandra.Batch.
type FakeBatch struct {
	mock.Mock
}

func (fb *FakeBatch) WithContext(ctx context.Context) Batch {
	fb.Called(ctx)
	return fb
}

func (fb *FakeBatch) Query(stmt string, values ...interface{}) {
	fb.Called(stmt, values)
}

// --- UNIT TESTS USING THE MOCKS ---

func TestInsertData(t *testing.T) {
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	// Prepare a fake batch.
	fakeBatch := new(FakeBatch)
	fakeBatch.On("Query", mock.Anything, mock.Anything).Times(2).Return()
	fakeBatch.On("WithContext", mock.Anything).Return(fakeBatch)
	mockSession.On("NewBatch", gocql.LoggedBatch).Return(fakeBatch)
	mockSession.On("ExecuteBatch", fakeBatch).Return(nil)

	// (Optionally, you can set expectations on fakeBatch.Query if you want to verify the queries added.)

	tripStartTime := time.Date(2023, 1, 15, 10, 0, 1, 0, time.UTC)
	carriage := &models.Carriage{
		TripID: models.TripID{
			RouteID:   "route_test",
			StartTime: tripStartTime,
		},
		EndTime:    tripStartTime.Add(1 * time.Hour),
		CarriageID: 10,
		Carts: []models.Cart{
			{
				CartID: models.CartID{
					EmployeeID:    "12345",
					OperationTime: time.Date(2023, 1, 15, 12, 30, 0, 0, time.UTC),
				},
				OperationType: 1,
				Items: []models.Item{
					{ProductID: 1, Quantity: 10, Price: 100},
				},
			},
		},
	}

	err := repo.InsertData(context.Background(), carriage)
	assert.NoError(t, err)
	mockSession.AssertExpectations(t)
}

func TestInsertData_ExecuteBatchError(t *testing.T) {
	// Create a mock session and repository.
	mockSession := new(MockSession)
	// Using a NopLogger which always returns nil on Log.
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	// Prepare a fake batch.
	fakeBatch := new(FakeBatch)
	// Expect WithContext to be called and return the same fake batch.
	fakeBatch.On("WithContext", mock.Anything).Return(fakeBatch)
	// In this test we have one cart with one employee trip insertion and one item insertion.
	// Therefore, we expect two calls to Query.
	fakeBatch.On("Query", mock.Anything, mock.Anything).Times(2).Return()

	// Set up the session so that when NewBatch is called it returns our fake batch.
	mockSession.On("NewBatch", gocql.LoggedBatch).Return(fakeBatch)
	// Simulate an error when ExecuteBatch is called.
	expectedErr := fmt.Errorf("batch error")
	mockSession.On("ExecuteBatch", fakeBatch).Return(expectedErr)

	// Create a dummy carriage report.
	tripStartTime := time.Date(2023, 1, 15, 10, 0, 1, 0, time.UTC)
	carriage := &models.Carriage{
		TripID: models.TripID{
			RouteID:   "route_test",
			StartTime: tripStartTime,
		},
		EndTime:    tripStartTime.Add(1 * time.Hour),
		CarriageID: 10,
		Carts: []models.Cart{
			{
				CartID: models.CartID{
					EmployeeID:    "12345",
					OperationTime: time.Date(2023, 1, 15, 12, 30, 0, 0, time.UTC),
				},
				OperationType: 1,
				Items: []models.Item{
					{ProductID: 1, Quantity: 10, Price: 100},
				},
			},
		},
	}

	// Call InsertData which should now hit the error branch.
	err := repo.InsertData(context.Background(), carriage)
	assert.Error(t, err, "expected error from ExecuteBatch")
	// Since we are using a NopLogger, Log returns nil; the error should be wrapped with "failed to execute batch:".
	assert.Contains(t, err.Error(), "failed to execute batch:")
	assert.Contains(t, err.Error(), "batch error")

	mockSession.AssertExpectations(t)
	fakeBatch.AssertExpectations(t)
}

func TestGetEmployeeCartsInTrip(t *testing.T) {
	// Create a mock session.
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	// Create a FakeQuery and set expectations.
	fakeQuery := new(FakeQuery)
	// Expect WithContext to be called and return fakeQuery.
	fakeQuery.On("WithContext", mock.Anything).Return(fakeQuery)
	// Prepare a simple fake iterator with one row.
	simpleIter := &SimpleFakeIter{
		rows: []fakeRow{
			{
				operationTime: time.Date(2023, 1, 15, 12, 30, 0, 0, time.UTC),
				operationType: 1,
				productID:     1,
				quantity:      10,
				price:         100,
			},
			{
				operationTime: time.Date(2023, 1, 15, 12, 30, 0, 0, time.UTC),
				operationType: 1,
				productID:     4,
				quantity:      10,
				price:         100,
			},
		},

		index: 0,
	}
	// Expect Iter() to be called on fakeQuery and return our simple iterator.
	fakeQuery.On("Iter").Return(simpleIter)

	// Expect Query to be called on the session.
	mockSession.On("Query", mock.Anything, mock.Anything).Return(fakeQuery)

	// Define test tripID and employeeID.
	tripID := &models.TripID{
		RouteID:   "route_test",
		Year:      "2023",
		StartTime: time.Date(2023, 1, 15, 10, 0, 1, 0, time.UTC),
	}
	employeeID := "testEmp"

	// Call the method under test.
	carts, err := repo.GetEmployeeCartsInTrip(context.Background(), tripID, &employeeID)
	assert.NoError(t, err)

	// Verify that two carts were aggregated.
	assert.Equal(t, 1, len(carts), "expected one cart")

	// Verify the details of the aggregated cart.
	cart := carts[0]
	assert.Equal(t, employeeID, cart.CartID.EmployeeID)
	assert.Equal(t, time.Date(2023, 1, 15, 12, 30, 0, 0, time.UTC), cart.CartID.OperationTime)
	assert.Equal(t, int8(1), cart.OperationType)
	assert.Equal(t, 2, len(cart.Items))
	item1 := cart.Items[0]
	assert.Equal(t, 1, item1.ProductID)
	assert.Equal(t, int16(10), item1.Quantity)
	assert.Equal(t, int64(100), item1.Price)
	item2 := cart.Items[1]
	assert.Equal(t, 4, item2.ProductID)
	assert.Equal(t, int16(10), item2.Quantity)
	assert.Equal(t, int64(100), item2.Price)

	// Verify expectations.
	mockSession.AssertExpectations(t)
	fakeQuery.AssertExpectations(t)
}

func TestGetEmployeeCartsInTrip_AggregateCloseError(t *testing.T) {
	// This test simulates an error during aggregation: aggregateCartsFromRows calls Close() and gets an error.
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	// Create a fake iterator that returns an error on its first Close() call.
	fakeIter := &fakeIterWithError{errorOnCall: 1} // first call returns error

	fakeQuery := new(FakeQuery)
	fakeQuery.On("WithContext", mock.Anything).Return(fakeQuery)
	fakeQuery.On("Iter").Return(fakeIter)

	// Expect Query to be called on the session.
	mockSession.On("Query", mock.Anything, mock.Anything).Return(fakeQuery)

	tripID := &models.TripID{
		RouteID:   "route_test",
		StartTime: time.Date(2023, 1, 15, 10, 0, 1, 0, time.UTC),
	}
	empID := "emp123"

	carts, err := repo.GetEmployeeCartsInTrip(context.Background(), tripID, &empID)
	assert.Nil(t, carts)
	assert.Error(t, err)
	// We expect the error from fakeIterWithError.
	assert.Contains(t, err.Error(), "iter close error")

	mockSession.AssertExpectations(t)
	fakeQuery.AssertExpectations(t)
}

func TestGetEmployeeCartsInTrip_PostAggregateCloseError(t *testing.T) {
	// This test simulates a successful aggregation (Close returns nil inside aggregateCartsFromRows)
	// but then GetEmployeeCartsInTrip calls Close again and that call returns an error.
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	// Create a fake iterator that returns nil on its first Close call and an error on the second.
	fakeIter := &fakeIterWithError{errorOnCall: 2} // first call nil, second call error

	fakeQuery := new(FakeQuery)
	fakeQuery.On("WithContext", mock.Anything).Return(fakeQuery)
	fakeQuery.On("Iter").Return(fakeIter)

	mockSession.On("Query", mock.Anything, mock.Anything).Return(fakeQuery)

	tripID := &models.TripID{
		RouteID:   "route_test",
		Year:      "2023",
		StartTime: time.Date(2023, 1, 15, 10, 0, 1, 0, time.UTC),
	}
	empID := "emp123"

	carts, err := repo.GetEmployeeCartsInTrip(context.Background(), tripID, &empID)
	assert.Nil(t, carts)
	assert.Error(t, err)
	// The error is from the second call to Close.
	assert.Contains(t, err.Error(), "iter close error")

	mockSession.AssertExpectations(t)
	fakeQuery.AssertExpectations(t)
}

func TestGetEmployeeIDsByTrip(t *testing.T) {
	// Create a mock session and repository.
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	// Create a FakeQuery and set expectations.
	fakeQuery := new(FakeQuery)
	fakeQuery.On("WithContext", mock.Anything).Return(fakeQuery)

	// Prepare a fake iterator that returns multiple employee IDs, including duplicates.
	fakeIter := &simpleFakeIterString{
		employeeIDs: []string{"emp1", "emp2", "emp1"},
		index:       0,
	}
	fakeQuery.On("Iter").Return(fakeIter)

	// Expect the repository to call Query on the session.
	mockSession.On("Query", mock.Anything, mock.Anything).Return(fakeQuery)

	// Define a test TripID.
	tripID := &models.TripID{
		RouteID:   "route_test",
		Year:      "2023",
		StartTime: time.Date(2023, 1, 15, 10, 0, 1, 0, time.UTC),
	}

	// Call GetEmployeeIDsByTrip.
	employeeIDs, err := repo.GetEmployeeIDsByTrip(context.Background(), tripID)
	assert.NoError(t, err)

	// Expect only unique IDs.
	expectedIDs := []string{"emp1", "emp2"}
	assert.ElementsMatch(t, expectedIDs, employeeIDs)

	mockSession.AssertExpectations(t)
	fakeQuery.AssertExpectations(t)
}

func TestGetEmployeeIDsByTrip_CloseError(t *testing.T) {
	// Create a mock session and repository.
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	// Create a FakeQuery and set expectations.
	fakeQuery := new(FakeQuery)
	fakeQuery.On("WithContext", mock.Anything).Return(fakeQuery)
	// Return a fake iterator that produces some rows but returns an error on Close().
	fakeIter := &fakeIterStringWithCloseError{
		employeeIDs: []string{"emp1", "emp2"},
		index:       0,
	}
	fakeQuery.On("Iter").Return(fakeIter)

	// Expect Query to be called on the mock session.
	mockSession.On("Query", mock.Anything, mock.Anything).Return(fakeQuery)

	// Define a test TripID.
	tripID := &models.TripID{
		RouteID:   "route_test",
		Year:      "2023",
		StartTime: time.Date(2023, 1, 15, 10, 0, 1, 0, time.UTC),
	}

	// Call GetEmployeeIDsByTrip, which should trigger the Close() error.
	employeeIDs, err := repo.GetEmployeeIDsByTrip(context.Background(), tripID)
	assert.Nil(t, employeeIDs)
	assert.Error(t, err)
	// Assert that the error message contains our simulated error text.
	assert.Contains(t, err.Error(), "close error")

	mockSession.AssertExpectations(t)
	fakeQuery.AssertExpectations(t)
}

func TestGetEmployeeTrips(t *testing.T) {
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	// Set up a FakeQuery and expectations.
	fakeQuery := new(FakeQuery)
	fakeQuery.On("WithContext", mock.Anything).Return(fakeQuery)
	// Prepare a fake iterator with two rows including the year field.
	iterRows := []struct {
		routeID   string
		year      string
		startTime time.Time
		endTime   time.Time
	}{
		{
			routeID:   "route1",
			year:      "2023",
			startTime: time.Date(2023, 1, 15, 10, 0, 0, 0, time.UTC),
			endTime:   time.Date(2023, 1, 15, 11, 0, 0, 0, time.UTC),
		},
		{
			routeID:   "route2",
			year:      "2023",
			startTime: time.Date(2023, 1, 16, 12, 0, 0, 0, time.UTC),
			endTime:   time.Date(2023, 1, 16, 13, 0, 0, 0, time.UTC),
		},
	}
	fakeIter := &simpleFakeIterTrip{rows: iterRows, index: 0}
	fakeQuery.On("Iter").Return(fakeIter)

	// Expect Query to be called on the mock session.
	mockSession.On("Query", mock.Anything, mock.Anything).Return(fakeQuery)

	employeeID := "emp123"
	year := "2023"

	trips, err := repo.GetEmployeeTrips(context.Background(), employeeID, year)
	assert.NoError(t, err)

	// Expected trips based on our fake iterator.
	expectedTrips := []models.EmployeeTrip{
		{
			EmployeeID: employeeID,
			TripID: models.TripID{
				RouteID:   "route1",
				Year:      year,
				StartTime: time.Date(2023, 1, 15, 10, 0, 0, 0, time.UTC),
			},
			EndTime: time.Date(2023, 1, 15, 11, 0, 0, 0, time.UTC),
		},
		{
			EmployeeID: employeeID,
			TripID: models.TripID{
				RouteID:   "route2",
				Year:      year,
				StartTime: time.Date(2023, 1, 16, 12, 0, 0, 0, time.UTC),
			},
			EndTime: time.Date(2023, 1, 16, 13, 0, 0, 0, time.UTC),
		},
	}

	assert.Equal(t, expectedTrips, trips)

	mockSession.AssertExpectations(t)
	fakeQuery.AssertExpectations(t)
}

func TestGetEmployeeTrips_CloseError(t *testing.T) {
	// Create a mock session and repository.
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	// Create a fake iterator that simulates a Close() error.
	fakeIter := &fakeIterTripWithCloseError{}

	// Create a fake query and set expectations.
	fakeQuery := new(FakeQuery)
	fakeQuery.On("WithContext", mock.Anything).Return(fakeQuery)
	fakeQuery.On("Iter").Return(fakeIter)

	// Expect Query to be called on the mock session.
	mockSession.On("Query", mock.Anything, mock.Anything).Return(fakeQuery)

	employeeID := "emp123"
	year := "2023"
	trips, err := repo.GetEmployeeTrips(context.Background(), employeeID, year)
	assert.Nil(t, trips)
	assert.Error(t, err)
	// Verify that the error message contains our simulated error text.
	assert.Contains(t, err.Error(), "trip iter close error")

	mockSession.AssertExpectations(t)
	fakeQuery.AssertExpectations(t)
}

func TestUpdateItemQuantity(t *testing.T) {
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	tripID := &models.TripID{
		RouteID:   "route_test",
		StartTime: time.Date(2023, 1, 15, 10, 0, 1, 0, time.UTC),
	}
	cartID := &models.CartID{
		EmployeeID:    "12345",
		OperationTime: time.Date(2023, 1, 15, 12, 30, 0, 0, time.UTC),
	}
	productID := 1
	newQuantity := int16(15)

	// Create a fake query and set expectations for ScanCAS.
	fakeQuery := new(FakeQuery)
	fakeQuery.On("ScanCAS", mock.Anything).Return(true, nil)
	// For WithContext, simply return the fakeQuery itself.
	fakeQuery.On("WithContext", mock.Anything).Return(fakeQuery)

	// Expect Query to be called with a statement matching the update query.
	updateQueryMatcher := mock.MatchedBy(func(stmt string) bool {
		return stmt == updateItemQuantityQuery || len(stmt) > 0 && stmt[0:6] == "UPDATE"
	})
	mockSession.On("Query", updateQueryMatcher, mock.Anything).Return(fakeQuery)

	err := repo.UpdateItemQuantity(context.Background(), tripID, cartID, &productID, &newQuantity)
	assert.NoError(t, err)
	mockSession.AssertExpectations(t)
	fakeQuery.AssertExpectations(t)
}

func TestUpdateItemQuantity_ScanCASError(t *testing.T) {
	// Create a mock session and repository.
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	// Define dummy parameters.
	tripID := &models.TripID{
		RouteID:   "route_test",
		Year:      "2023",
		StartTime: time.Date(2023, 1, 15, 10, 0, 0, 0, time.UTC),
	}
	cartID := &models.CartID{
		EmployeeID:    "12345",
		OperationTime: time.Date(2023, 1, 15, 12, 30, 0, 0, time.UTC),
	}
	productID := 1
	newQuantity := int16(15)

	// Set up a fake query.
	fakeQuery := new(FakeQuery)
	// When WithContext is called, return the fakeQuery.
	fakeQuery.On("WithContext", mock.Anything).Return(fakeQuery)
	// Simulate ScanCAS returning an error.
	scanErr := fmt.Errorf("scan error")
	fakeQuery.On("ScanCAS", mock.Anything).Return(false, scanErr)

	// Expect Query to be called on the session.
	mockSession.On("Query", mock.Anything, mock.Anything).Return(fakeQuery)

	// Call the method under test.
	err := repo.UpdateItemQuantity(context.Background(), tripID, cartID, &productID, &newQuantity)
	// Assert that an error is returned and it matches the simulated error.
	assert.Error(t, err)
	assert.Equal(t, scanErr, err)

	mockSession.AssertExpectations(t)
	fakeQuery.AssertExpectations(t)
}

func TestUpdateItemQuantity_NotApplied(t *testing.T) {
	// Create a mock session and repository.
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	// Define dummy parameters.
	tripID := &models.TripID{
		RouteID:   "route_test",
		Year:      "2023",
		StartTime: time.Date(2023, 1, 15, 10, 0, 0, 0, time.UTC),
	}
	cartID := &models.CartID{
		EmployeeID:    "12345",
		OperationTime: time.Date(2023, 1, 15, 12, 30, 0, 0, time.UTC),
	}
	productID := 1
	newQuantity := int16(15)

	// Set up a fake query.
	fakeQuery := new(FakeQuery)
	fakeQuery.On("WithContext", mock.Anything).Return(fakeQuery)
	// Simulate ScanCAS returning (false, nil) indicating that the update was not applied.
	fakeQuery.On("ScanCAS", mock.Anything).Return(false, nil)

	// Expect Query to be called on the session.
	mockSession.On("Query", mock.Anything, mock.Anything).Return(fakeQuery)

	// Call the method under test.
	err := repo.UpdateItemQuantity(context.Background(), tripID, cartID, &productID, &newQuantity)
	// Assert that an error is returned with the message "transaction does not exist".
	assert.Error(t, err)
	assert.EqualError(t, err, "transaction does not exist")

	mockSession.AssertExpectations(t)
	fakeQuery.AssertExpectations(t)
}

func TestDeleteItemFromCart(t *testing.T) {
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	tripID := &models.TripID{
		RouteID:   "route_test",
		Year:      "2023",
		StartTime: time.Date(2023, 1, 15, 10, 0, 1, 0, time.UTC),
	}
	cartID := &models.CartID{
		EmployeeID:    "12345",
		OperationTime: time.Date(2023, 1, 15, 12, 30, 0, 0, time.UTC),
	}
	productID := 1

	fakeQuery := new(FakeQuery)
	fakeQuery.On("ScanCAS", mock.Anything).Return(true, nil)
	fakeQuery.On("WithContext", mock.Anything).Return(fakeQuery)

	deleteQueryMatcher := mock.MatchedBy(func(stmt string) bool {
		return stmt == deleteItemFromCartQuery || len(stmt) > 0 && stmt[0:6] == "DELETE"
	})
	mockSession.On("Query", deleteQueryMatcher, mock.Anything).Return(fakeQuery)

	err := repo.DeleteItemFromCart(context.Background(), tripID, cartID, &productID)
	assert.NoError(t, err)
	mockSession.AssertExpectations(t)
	fakeQuery.AssertExpectations(t)
}

func TestDeleteItemFromCart_ScanCASError(t *testing.T) {
	// This test simulates an error from ScanCAS.
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	tripID := &models.TripID{
		RouteID:   "route_test",
		Year:      "2023",
		StartTime: time.Date(2023, 1, 15, 10, 0, 0, 0, time.UTC),
	}
	cartID := &models.CartID{
		EmployeeID:    "12345",
		OperationTime: time.Date(2023, 1, 15, 12, 30, 0, 0, time.UTC),
	}
	productID := 1

	// Create a fake query and set expectations.
	fakeQuery := new(FakeQuery)
	fakeQuery.On("WithContext", mock.Anything).Return(fakeQuery)
	// Simulate ScanCAS returning an error.
	scanErr := fmt.Errorf("scan error")
	fakeQuery.On("ScanCAS", mock.Anything).Return(false, scanErr)

	// Expect Query to be called on the session.
	mockSession.On("Query", mock.Anything, mock.Anything).Return(fakeQuery)

	// Call DeleteItemFromCart, which should hit the error branch.
	err := repo.DeleteItemFromCart(context.Background(), tripID, cartID, &productID)
	assert.Error(t, err)
	assert.Equal(t, scanErr, err)

	mockSession.AssertExpectations(t)
	fakeQuery.AssertExpectations(t)
}

func TestDeleteItemFromCart_NotDeleted(t *testing.T) {
	// This test simulates a case where ScanCAS succeeds but returns false,
	// so the deletion is not applied.
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	tripID := &models.TripID{
		RouteID:   "route_test",
		Year:      "2023",
		StartTime: time.Date(2023, 1, 15, 10, 0, 0, 0, time.UTC),
	}
	cartID := &models.CartID{
		EmployeeID:    "12345",
		OperationTime: time.Date(2023, 1, 15, 12, 30, 0, 0, time.UTC),
	}
	productID := 1

	fakeQuery := new(FakeQuery)
	fakeQuery.On("WithContext", mock.Anything).Return(fakeQuery)
	// Simulate ScanCAS returning (false, nil) to indicate the item was not deleted.
	fakeQuery.On("ScanCAS", mock.Anything).Return(false, nil)

	mockSession.On("Query", mock.Anything, mock.Anything).Return(fakeQuery)

	err := repo.DeleteItemFromCart(context.Background(), tripID, cartID, &productID)
	assert.Error(t, err)
	assert.EqualError(t, err, "item does not exist")

	mockSession.AssertExpectations(t)
	fakeQuery.AssertExpectations(t)
}
