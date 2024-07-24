package cassandra

import (
	"github.com/gocql/gocql"
)

type SalesRepository struct {
	session *gocql.Session
}
