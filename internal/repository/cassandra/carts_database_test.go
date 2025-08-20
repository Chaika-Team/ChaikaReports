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

// --- Used in TestGetTrip ---

type tripOpRow struct {
	routeID       string
	startTime     time.Time
	employeeID    string
	operationTime time.Time
	productID     int
	carriageID    int8
	endTime       time.Time
	operationType int8
	price         int64
	quantity      int16
}

type fakeTripIter struct {
	rows     []tripOpRow
	index    int
	closeErr error
}

func (f *fakeTripIter) Scan(dest ...interface{}) bool {
	if f.index >= len(f.rows) {
		return false
	}
	r := f.rows[f.index]
	f.index++

	// 9 columns
	if len(dest) != 10 {
		return false
	}

	*dest[0].(*string) = r.routeID
	*dest[1].(*time.Time) = r.startTime
	*dest[2].(*string) = r.employeeID
	*dest[3].(*time.Time) = r.operationTime
	*dest[4].(*int) = r.productID
	*dest[5].(*int8) = r.carriageID
	*dest[6].(*time.Time) = r.endTime
	*dest[7].(*int8) = r.operationType
	*dest[8].(*int64) = r.price
	*dest[9].(*int16) = r.quantity
	return true
}

func (f *fakeTripIter) Close() error {
	return f.closeErr
}

// --- Used in TestGetUnsyncedTrips ---
type unsyncRow struct {
	routeID   string
	startTime time.Time
	year      string
}

type fakeUnsyncIter struct {
	rows  []unsyncRow
	index int
	err   error
}

func (f *fakeUnsyncIter) Scan(dest ...interface{}) bool {
	if f.index >= len(f.rows) {
		return false
	}
	r := f.rows[f.index]
	f.index++
	if len(dest) != 3 {
		return false
	}
	*dest[0].(*string) = r.routeID
	*dest[1].(*time.Time) = r.startTime
	*dest[2].(*string) = r.year
	return true
}

func (f *fakeUnsyncIter) Close() error { return f.err }

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

func (fq *FakeQuery) PageSize(n int) Query {
	return fq
}
func (fq *FakeQuery) PageState(state []byte) Query {
	return fq
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

func (fi *FakeIter) PageState() []byte {
	return nil
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
	fakeBatch.On("Query", mock.Anything, mock.Anything).Times(4).Return()
	fakeBatch.On("WithContext", mock.Anything).Return(fakeBatch)
	mockSession.On("NewBatch", gocql.LoggedBatch).Return(fakeBatch)
	mockSession.On("ExecuteBatch", fakeBatch).Return(nil)

	// (Optionally, you can set expectations on fakeBatch.Query if you want to verify the queries added.)

	tripStartTime := time.Date(2023, 1, 15, 10, 0, 1, 0, time.UTC)
	carriage := &models.CarriageReport{
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
	fakeBatch.On("Query", mock.Anything, mock.Anything).Times(4).Return()

	// Set up the session so that when NewBatch is called it returns our fake batch.
	mockSession.On("NewBatch", gocql.LoggedBatch).Return(fakeBatch)
	// Simulate an error when ExecuteBatch is called.
	expectedErr := fmt.Errorf("batch error")
	mockSession.On("ExecuteBatch", fakeBatch).Return(expectedErr)

	// Create a dummy carriage report.
	tripStartTime := time.Date(2023, 1, 15, 10, 0, 1, 0, time.UTC)
	carriage := &models.CarriageReport{
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

	fakeIter := &fakeIterWithError{errorOnCall: 1}

	fakeQuery := new(FakeQuery)
	fakeQuery.On("WithContext", mock.Anything).Return(fakeQuery)
	fakeQuery.On("Iter").Return(fakeIter)

	// Expect Query to be called on the session.
	mockSession.On("Query", mock.Anything, mock.Anything).Return(fakeQuery)

	tripID := &models.TripID{
		RouteID:   "route_test",
		Year:      "year",
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

func (s *SimpleFakeIter) PageState() []byte               { return nil }
func (s *simpleFakeIterString) PageState() []byte         { return nil }
func (s *simpleFakeIterTrip) PageState() []byte           { return nil }
func (f *fakeIterWithError) PageState() []byte            { return nil }
func (f *fakeIterStringWithCloseError) PageState() []byte { return nil }
func (f *fakeIterTripWithCloseError) PageState() []byte   { return nil }
func (f *fakeTripIter) PageState() []byte                 { return nil }
func (f *fakeUnsyncIter) PageState() []byte               { return nil }

func TestGetEmployeeCartsInTripPaged_FirstPage_WithLimit_ReturnsTwoCarts_AndCursor(t *testing.T) {
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	// Times
	start := time.Date(2025, 8, 20, 8, 0, 0, 0, time.UTC)
	op1 := time.Date(2025, 8, 20, 10, 0, 0, 0, time.UTC) // cart A
	op2 := time.Date(2025, 8, 20, 9, 0, 0, 0, time.UTC)  // cart B
	op3 := time.Date(2025, 8, 20, 8, 30, 0, 0, time.UTC) // cart C

	// First page query iterator (no cursor): rows for 3 carts (we will limit by carts=2)
	iter1 := &SimpleFakeIter{
		rows: []fakeRow{
			// Cart A (2 items)
			{operationTime: op1, operationType: 1, productID: 1, quantity: 2, price: 100},
			{operationTime: op1, operationType: 1, productID: 2, quantity: 1, price: 200},

			// Cart B (1 item)
			{operationTime: op2, operationType: 1, productID: 3, quantity: 1, price: 300},

			// Cart C (1 item) – should not be emitted because cartLimit=2
			{operationTime: op3, operationType: 1, productID: 4, quantity: 1, price: 400},
		},
	}

	q1 := new(FakeQuery)
	q1.On("WithContext", mock.Anything).Return(q1)
	q1.On("Iter").Return(iter1)

	// Expect base query to be used (no cursor)
	mockSession.On("Query", getEmployeeCartsInTripQuery, mock.Anything).Return(q1)

	tripID := &models.TripID{RouteID: "routeX", Year: "2025", StartTime: start}
	emp := "emp1"

	carts, next, err := repo.GetEmployeeCartsInTripPaged(context.Background(), tripID, emp, 2, "")
	assert.NoError(t, err)

	// Should return 2 complete carts (op1 and op2)
	require := assert.New(t)
	require.Equal(2, len(carts))

	// First emitted cart is op1 (order doesn’t strictly matter for the test logic, but check contents)
	var foundOp1, foundOp2 bool
	for _, c := range carts {
		if c.CartID.OperationTime.Equal(op1) {
			foundOp1 = true
			assert.Equal(t, "emp1", c.CartID.EmployeeID)
			assert.Equal(t, int8(1), c.OperationType)
			assert.Len(t, c.Items, 2)
			assert.Equal(t, 1, c.Items[0].ProductID)
			assert.Equal(t, 2, int(c.Items[0].Quantity))
			assert.Equal(t, int64(100), c.Items[0].Price)
			assert.Equal(t, 2, c.Items[1].ProductID)
		}
		if c.CartID.OperationTime.Equal(op2) {
			foundOp2 = true
			assert.Len(t, c.Items, 1)
			assert.Equal(t, 3, c.Items[0].ProductID)
		}
	}
	require.True(foundOp1)
	require.True(foundOp2)

	// We should get a non-empty cursor based on last emitted cart (op2)
	assert.NotEmpty(t, next)

	mockSession.AssertExpectations(t)
	q1.AssertExpectations(t)
}

func TestGetEmployeeCartsInTripPaged_NextPage_WithCursor_ReturnsRemaining_NoCursor(t *testing.T) {
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	start := time.Date(2025, 8, 20, 8, 0, 0, 0, time.UTC)
	emp := "emp1"

	// We need a first call to produce a cursor so we can reuse it here.
	// But since cursor format is opaque base64, we can just call the repo once quickly
	// to get a real cursor and then stub the "after" query.
	// (Alternatively, if you know your encodeCursor format you can hardcode a value.)

	// --- First call to get a cursor (limit=1, with two carts) ---
	op1 := time.Date(2025, 8, 20, 10, 0, 0, 0, time.UTC)
	op2 := time.Date(2025, 8, 20, 9, 0, 0, 0, time.UTC)

	iter1 := &SimpleFakeIter{
		rows: []fakeRow{
			{operationTime: op1, operationType: 1, productID: 1, quantity: 1, price: 100},
			{operationTime: op2, operationType: 1, productID: 2, quantity: 1, price: 200},
		},
	}
	q1 := new(FakeQuery)
	q1.On("WithContext", mock.Anything).Return(q1)
	q1.On("Iter").Return(iter1)
	mockSession.On("Query", getEmployeeCartsInTripQuery, mock.Anything).Return(q1)

	tripID := &models.TripID{RouteID: "routeX", Year: "2025", StartTime: start}

	firstPage, cursor, err := repo.GetEmployeeCartsInTripPaged(context.Background(), tripID, emp, 1, "")
	assert.NoError(t, err)
	assert.Len(t, firstPage, 1)
	assert.NotEmpty(t, cursor)
	mockSession.AssertExpectations(t)
	q1.AssertExpectations(t)

	// --- Second call: use the returned cursor (should return the remaining cart and empty cursor) ---
	iter2 := &SimpleFakeIter{
		rows: []fakeRow{
			{operationTime: op2, operationType: 1, productID: 2, quantity: 1, price: 200},
		},
	}
	q2 := new(FakeQuery)
	q2.On("WithContext", mock.Anything).Return(q2)
	q2.On("Iter").Return(iter2)
	mockSession.ExpectedCalls = nil // reset expectations for second call
	mockSession.On("Query", getEmployeeCartsInTripAfterCursorQuery, mock.Anything).Return(q2)

	secondPage, next, err := repo.GetEmployeeCartsInTripPaged(context.Background(), tripID, emp, 2, cursor)
	assert.NoError(t, err)
	assert.Len(t, secondPage, 1)
	assert.Empty(t, next)

	mockSession.AssertExpectations(t)
	q2.AssertExpectations(t)
}

func TestGetEmployeeCartsInTripPaged_InvalidCursor_ReturnsError(t *testing.T) {
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	start := time.Date(2025, 8, 20, 8, 0, 0, 0, time.UTC)
	tripID := &models.TripID{RouteID: "routeX", Year: "2025", StartTime: start}

	// Pass junk base64 to trigger decode error
	_, _, err := repo.GetEmployeeCartsInTripPaged(context.Background(), tripID, "emp1", 2, "!!!not-base64!!!")
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid cursor")

	// No expectations on session because we shouldn't even hit the DB
	mockSession.AssertExpectations(t)
}

func TestGetEmployeeCartsInTripPaged_NoLimit_ReturnsAll_NoCursor(t *testing.T) {
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	start := time.Date(2025, 8, 20, 8, 0, 0, 0, time.UTC)
	op1 := time.Date(2025, 8, 20, 10, 0, 0, 0, time.UTC)
	op2 := time.Date(2025, 8, 20, 9, 0, 0, 0, time.UTC)

	// cartLimit <= 0 means "no cutoff" in the repo implementation
	iter := &SimpleFakeIter{
		rows: []fakeRow{
			{operationTime: op1, operationType: 1, productID: 1, quantity: 1, price: 100},
			{operationTime: op1, operationType: 1, productID: 2, quantity: 1, price: 200},
			{operationTime: op2, operationType: 1, productID: 3, quantity: 1, price: 300},
		},
	}
	q := new(FakeQuery)
	q.On("WithContext", mock.Anything).Return(q)
	q.On("Iter").Return(iter)
	mockSession.On("Query", getEmployeeCartsInTripQuery, mock.Anything).Return(q)

	tripID := &models.TripID{RouteID: "routeX", Year: "2025", StartTime: start}

	carts, next, err := repo.GetEmployeeCartsInTripPaged(context.Background(), tripID, "emp1", 0, "")
	assert.NoError(t, err)
	assert.Len(t, carts, 2) // two carts (op1, op2)
	assert.Empty(t, next)

	mockSession.AssertExpectations(t)
	q.AssertExpectations(t)
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

func TestGetTrip(t *testing.T) {
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	start := time.Date(2023, 1, 15, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)

	iter := &fakeTripIter{
		rows: []tripOpRow{
			// two items in same cart (same emp / opTime) in carriage 1
			{"r1", start, "empA", start.Add(30 * time.Minute), 1, 1, end, 0, 100, 2},
			{"r1", start, "empA", start.Add(30 * time.Minute), 2, 1, end, 0, 200, 5},
			// another cart in same carriage
			{"r1", start, "empB", start.Add(40 * time.Minute), 3, 1, end, 1, 150, 1},
			// carriage 2
			{"r1", start, "empA", start.Add(50 * time.Minute), 4, 2, end, 0, 50, 3},
		},
	}

	fakeQuery := new(FakeQuery)
	fakeQuery.On("WithContext", mock.Anything).Return(fakeQuery)
	fakeQuery.On("Iter").Return(iter)
	mockSession.
		On("Query", mock.Anything, mock.Anything).
		Return(fakeQuery)

	tripID := &models.TripID{RouteID: "r1", Year: "2023", StartTime: start}

	got, err := repo.GetTrip(context.Background(), tripID)
	assert.NoError(t, err)

	// expect two carriages
	assert.Len(t, got.Carriage, 2)

	// carriage 1: two carts, first cart has two items
	var c1 models.CarriageReport
	for _, c := range got.Carriage {
		if c.CarriageID == 1 {
			c1 = c
			break
		}
	}
	assert.Len(t, c1.Carts, 2)

	for _, cart := range c1.Carts {
		if cart.CartID.EmployeeID == "empA" {
			assert.Len(t, cart.Items, 2)
		}
	}

	mockSession.AssertExpectations(t)
	fakeQuery.AssertExpectations(t)
}

/* ----------------------- GetTrip: iterator.Close error -------------------- */

func TestGetTrip_CloseError(t *testing.T) {
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	iter := &fakeTripIter{closeErr: fmt.Errorf("close boom")}
	fakeQuery := new(FakeQuery)
	fakeQuery.On("WithContext", mock.Anything).Return(fakeQuery)
	fakeQuery.On("Iter").Return(iter)
	mockSession.On("Query", mock.Anything, mock.Anything).Return(fakeQuery)

	tripID := &models.TripID{RouteID: "r1", Year: "2023", StartTime: time.Now()}
	_, err := repo.GetTrip(context.Background(), tripID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "close boom")
}

// -------------------

func TestGetUnsyncedTrips_Happy(t *testing.T) {
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	rows := []unsyncRow{{"r1", time.Now(), "2023"}}
	iter := &fakeUnsyncIter{rows: rows}
	fq := new(FakeQuery)
	fq.On("WithContext", mock.Anything).Return(fq)
	fq.On("Iter").Return(iter)
	mockSession.On("Query", getUnsyncedTripsQuery, mock.Anything).Return(fq)

	res, err := repo.GetUnsyncedTrips(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res))
}

func TestGetUnsyncedTrips_Empty(t *testing.T) {
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	iter := &fakeUnsyncIter{} // no rows
	fq := new(FakeQuery)
	fq.On("WithContext", mock.Anything).Return(fq)
	fq.On("Iter").Return(iter)
	mockSession.On("Query", getUnsyncedTripsQuery, mock.Anything).Return(fq)

	res, err := repo.GetUnsyncedTrips(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, res)
}

func TestGetUnsyncedTrips_CloseError(t *testing.T) {
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	iter := &fakeUnsyncIter{err: fmt.Errorf("iter close")}
	fq := new(FakeQuery)
	fq.On("WithContext", mock.Anything).Return(fq)
	fq.On("Iter").Return(iter)
	mockSession.On("Query", getUnsyncedTripsQuery, mock.Anything).Return(fq)

	_, err := repo.GetUnsyncedTrips(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "iter close")
}

func TestDeleteSyncedTrip_Success(t *testing.T) {
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	fq := new(FakeQuery)
	fq.On("WithContext", mock.Anything).Return(fq)
	fq.On("ScanCAS", mock.Anything).Return(true, nil)
	mockSession.On("Query", deleteTripFromUnsynchronizedTripsQuery, mock.Anything).Return(fq)

	err := repo.DeleteSyncedTrip(context.Background(), "r1", time.Now())
	assert.NoError(t, err)
	mockSession.AssertExpectations(t)
}

func TestDeleteSyncedTrip_NotExists(t *testing.T) {
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	fq := new(FakeQuery)
	fq.On("WithContext", mock.Anything).Return(fq)
	fq.On("ScanCAS", mock.Anything).Return(false, nil)
	mockSession.On("Query", deleteTripFromUnsynchronizedTripsQuery, mock.Anything, mock.Anything).
		Return(fq)

	err := repo.DeleteSyncedTrip(context.Background(), "r1", time.Now())
	assert.Error(t, err)
	assert.EqualError(t, err, "trip does not exist")
}

func TestDeleteSyncedTrip_ScanError(t *testing.T) {
	mockSession := new(MockSession)
	repo := NewSalesRepository(mockSession, log.NewNopLogger())

	scanErr := fmt.Errorf("scan err")
	fq := new(FakeQuery)
	fq.On("WithContext", mock.Anything).Return(fq)
	fq.On("ScanCAS", mock.Anything).Return(false, scanErr)
	mockSession.On("Query", deleteTripFromUnsynchronizedTripsQuery, mock.Anything, mock.Anything).
		Return(fq)

	err := repo.DeleteSyncedTrip(context.Background(), "r1", time.Now())
	assert.Error(t, err)
	assert.Equal(t, scanErr, err)
}
