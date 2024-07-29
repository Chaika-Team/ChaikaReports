package cassandra

import (
	"github.com/gocql/gocql"
	"log"
)

var Session *gocql.Session

func InitCassandra(keyspace string, hosts []string, username, password string) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: username,
		Password: password,
	}

	var err error
	Session, err = cluster.CreateSession()
	if err != nil {
		log.Fatalf("Failed to connect to the cluster: %v", err)
	}
	log.Println("Cassandra connection established")
}

func CloseCassandra() {
	Session.Close()
}
