package cassandra

import (
	"github.com/gocql/gocql"
	"log"
)

func InitCassandra(keyspace string, hosts []string, username, password string) (*gocql.Session, error) {
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
		log.Fatalf("Failed to connect to the cluster: %v", err)
		return nil, err
	}
	log.Println("Cassandra connection established")
	return session, nil

}

func CloseCassandra(session *gocql.Session) {
	session.Close()
}
