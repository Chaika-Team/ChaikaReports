package test

import (
	"ChaikaReports/internal/config"
	"ChaikaReports/internal/repository/cassandra"
	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CassandraTestSuite struct {
	suite.Suite
	testSession *gocql.Session
}

// SetupSuite is called before any tests are run. It sets up the keyspace, table, and connections.
func (suite *CassandraTestSuite) SetupSuite() {
	// Load the configuration
	cfg := config.LoadConfig()

	// Create the test keyspace
	testSession, err := cassandra.InitCassandra(cfg.CassandraTest.Keyspace, cfg.CassandraTest.Hosts, cfg.CassandraTest.User, cfg.CassandraTest.Password)
	assert.NoError(suite.T(), err, "Failed to connect to test keyspace")
	suite.testSession = testSession

	// Create identical table, because Cassandra doesn't have tools to copy table schemas :))))
	err = testSession.Query(`CREATE TABLE operations (
    	route_id text,
    	start_time timestamp,
    	end_time timestamp,
    	carriage_id tinyint,
    	employee_id text,
    	operation_type tinyint,
    	operation_time timestamp,
    	product_id int,
    	quantity smallint,
    	price float,
    	PRIMARY KEY (route_id, start_time, employee_id, operation_time, product_id))
    	WITH CLUSTERING ORDER BY (start_time DESC, employee_id ASC, operation_time DESC, product_id ASC)`).Exec()
	assert.NoError(suite.T(), err, "Failed to create table in keyspace")
}

func (suite *CassandraTestSuite) TearDownSuite() {
	// Drop the table after testing
	err := suite.testSession.Query(`DROP TABLE IF EXISTS operations`).Exec()
	assert.NoError(suite.T(), err, "Failed to drop test table")

	// Close the session
	cassandra.CloseCassandra(suite.testSession)
}
