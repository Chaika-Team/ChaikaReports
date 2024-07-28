package cassandra

import (
	"ChaikaReports/internal/config"
	"context"
	"fmt"
	"github.com/gocql/gocql"
	"log"
	"time"
)

type Client interface {
	// Exec executes a query without returning any rows.
	Exec(ctx context.Context, sql string, arguments ...interface{}) error

	// Query sends a query to the database and returns the rows.
	Query(ctx context.Context, sql string, args ...interface{}) (*gocql.Iter, error)

	// QueryRow sends a query to the database and returns a single row.
	QueryRow(ctx context.Context, sql string, args ...interface{}) *gocql.Query

	// SendBatch executes a batch of CQL statements.
	ExecuteBatch(batch *gocql.Batch) error
}

type cassandraClient struct {
	session *gocql.Session
}

func (c *cassandraClient) Exec(ctx context.Context, sql string, arguments ...interface{}) error {
	return c.session.Query(sql, arguments...).WithContext(ctx).Exec()
}

func (c *cassandraClient) Query(ctx context.Context, sql string, args ...interface{}) (*gocql.Iter, error) {
	return c.session.Query(sql, args...).WithContext(ctx).Iter(), nil
}

func (c *cassandraClient) QueryRow(ctx context.Context, sql string, args ...interface{}) *gocql.Query {
	return c.session.Query(sql, args...).WithContext(ctx)
}

func (c *cassandraClient) ExecuteBatch(batch *gocql.Batch) error {
	return c.session.ExecuteBatch(batch)
}

// NewClient creates a new Cassandra client.
func NewClient(ctx context.Context, config config.StorageConfig, maxAttempts int) (Client, error) {
	cluster := gocql.NewCluster(config.Hosts...)
	cluster.Keyspace = config.Keyspace
	cluster.Consistency = gocql.Quorum
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: config.User,
		Password: config.Password,
	}

	var session *gocql.Session
	var err error
	for i := 1; i <= maxAttempts; i++ {
		session, err = cluster.CreateSession()
		if err == nil {
			log.Println("Cassandra connection established")
			return &cassandraClient{session: session}, nil
		}
		if i < maxAttempts {
			log.Printf("Failed to connect to database, attempt %d/%d: %v", i, maxAttempts, err)
			time.Sleep(5 * time.Second)
		}
	}
	return nil, fmt.Errorf("failed to connect to the database after %d attempts: %v", maxAttempts, err)
}
