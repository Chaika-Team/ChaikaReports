package cassandra

import (
	"ChaikaReports/internal/models"
	"github.com/gocql/gocql"
)

type SalesRepository struct {
	session *gocql.Session
}

func (s *SalesRepository) StoreSalesData(salesData models.SalesData) error {
	return nil
}
