package cassandra

import (
	"github.com/go-kit/log"
	"github.com/gocql/gocql"
	"time"
)

func InitCassandra(logger log.Logger, keyspace string, hosts []string, username, password string, timeout time.Duration, delay time.Duration, attempts int) (CassandraSession, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = timeout
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: username,
		Password: password,
	}
	var session *gocql.Session
	var err error
	for i := 0; i < attempts; i++ {
		session, err = cluster.CreateSession()
		if err == nil {
			break
		}
		_ = logger.Log("msg", "Failed to create session, retrying", "attempt", i+1, "error", err)
		time.Sleep(delay)
	}

	if err != nil {
		_ = logger.Log("msg", "Failed to initialize Cassandra", "error", err)
		return nil, err
	}

	if err := session.Query("SELECT now() FROM system.local").Exec(); err != nil {
		_ = logger.Log("msg", "Cassandra health check failed", "error", err)
		session.Close()
		return nil, err
	}

	return &realSession{s: session}, nil
}

func CloseCassandra(session CassandraSession) {
	session.Close()
}
