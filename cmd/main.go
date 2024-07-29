package main

import (
	"ChaikaReports/internal/models"
	"ChaikaReports/internal/repository/cassandra"
	"fmt"
)
import "ChaikaReports/internal/config"

func main() {
	cnfg := config.LoadConfig()
	cassandra.InitCassandra(
		cnfg.Cassandra.Keyspace,
		cnfg.Cassandra.Hosts,
		cnfg.Cassandra.User,
		cnfg.Cassandra.Password,
	)

	r := cassandra.NewSalesRepository()
	salesData := &models.SalesData{
		TripID:      2,
		CarriageID:  101,
		ConductorID: 1003,
		Actions: []models.Action{
			{ProductID: 1, OperationTypeID: 1, Count: 10},
			{ProductID: 2, OperationTypeID: 2, Count: 20},
		},
	}
	fmt.Println(r.GetActionsByConductor(salesData.TripID, salesData.ConductorID))

}
