package cassandra

import (
	"github.com/go-kit/log"
	"github.com/gocql/gocql"
)

func InitCassandra(logger log.Logger, keyspace string, hosts []string, username, password string) (*gocql.Session, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: username,
		Password: password,
	}
	var session *gocql.Session
	var err error
	session, err = cluster.CreateSession()

	if err != nil {
		_ = logger.Log("msg", "Failed to initialize Cassandra", "error", err)
		return nil, err
	}

	// Defer closing of session to make sure it gets closed
	defer func() {
		if r := recover(); r != nil {
			_ = logger.Log("msg", "Recovered from panic during session", "error", r)
			session.Close()
		}
	}()

	// Performing simple health check for Cassandra DB connection
	if err := session.Query("SELECT now() FROM system.local").Exec(); err != nil {
		_ = logger.Log("msg", "Cassandra health check failed", "error", err)
		session.Close() // Close the session if the health check fails
		return nil, err
	}

	return session, nil
}

func CloseCassandra(session *gocql.Session) {
	session.Close()
}
