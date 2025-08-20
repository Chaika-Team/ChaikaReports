// cassandra/session.go
package cassandra

import (
	"context"
	"github.com/gocql/gocql"
)

// Query abstracts the methods of a gocql.Query used in the repository layer
type Query interface {
	WithContext(ctx context.Context) Query
	Exec() error
	Iter() Iter
	ScanCAS(dest ...interface{}) (bool, error)
	PageSize(n int) Query
	PageState(state []byte) Query
}

// Iter abstracts gocql.Iter used to iterate over query results
type Iter interface {
	Scan(dest ...interface{}) bool
	Close() error
	PageState() []byte
}

// Batch abstracts a batch of queries
type Batch interface {
	WithContext(ctx context.Context) Batch
	Query(stmt string, values ...interface{})
}

// CassandraSession abstracts the Cassandra session
type CassandraSession interface {
	Query(stmt string, values ...interface{}) Query
	NewBatch(batchType gocql.BatchType) Batch
	ExecuteBatch(batch Batch) error
	Close()
}

// --- Concrete wrappers around gocql types ---

// queryWrapper wraps a *gocql.Query to implement the Query interface
type queryWrapper struct {
	q *gocql.Query
}

func (qw *queryWrapper) WithContext(ctx context.Context) Query {
	return &queryWrapper{q: qw.q.WithContext(ctx)}
}

func (qw *queryWrapper) Exec() error {
	return qw.q.Exec()
}

func (qw *queryWrapper) Iter() Iter {
	return &iterWrapper{iter: qw.q.Iter()}
}

func (qw *queryWrapper) ScanCAS(dest ...interface{}) (bool, error) {
	return qw.q.ScanCAS(dest...)
}

func (qw *queryWrapper) PageSize(n int) Query {
	qw.q = qw.q.PageSize(n)
	return qw
}
func (qw *queryWrapper) PageState(state []byte) Query {
	qw.q = qw.q.PageState(state)
	return qw
}

// iterWrapper wraps a *gocql.Iter
type iterWrapper struct {
	iter *gocql.Iter
}

func (iw *iterWrapper) Scan(dest ...interface{}) bool {
	return iw.iter.Scan(dest...)
}

func (iw *iterWrapper) Close() error {
	return iw.iter.Close()
}

func (iw *iterWrapper) PageState() []byte {
	return iw.iter.PageState()
}

// batchWrapper wraps a *gocql.Batch to implement the Batch interface
type batchWrapper struct {
	b *gocql.Batch
}

func (bw *batchWrapper) WithContext(ctx context.Context) Batch {
	newBatch := bw.b.WithContext(ctx)
	return &batchWrapper{b: newBatch}
}

func (bw *batchWrapper) Query(stmt string, values ...interface{}) {
	bw.b.Query(stmt, values...)
}

// realSession wraps a *gocql.Session to implement CassandraSession
type realSession struct {
	s *gocql.Session
}

func (rs *realSession) Query(stmt string, values ...interface{}) Query {
	q := rs.s.Query(stmt, values...)
	return &queryWrapper{q: q}
}

func (rs *realSession) NewBatch(batchType gocql.BatchType) Batch {
	b := rs.s.NewBatch(batchType)
	return &batchWrapper{b: b}
}

func (rs *realSession) ExecuteBatch(batch Batch) error {
	if bw, ok := batch.(*batchWrapper); ok {
		return rs.s.ExecuteBatch(bw.b)
	}
	return nil
}

func (rs *realSession) Close() {
	rs.s.Close()
}
