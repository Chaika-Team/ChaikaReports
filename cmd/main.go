package main

import (
	"fmt"
	"github.com/gocql/gocql"
	"log"
	"time"
)

func main() {
	// Initialize the cluster
	cluster := gocql.NewCluster("188.242.205.5")
	cluster.Port = 9042
	cluster.Keyspace = "system"
	cluster.Consistency = gocql.Quorum

	// Add authentication if needed
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: "grisha",
		Password: "n*_irR#2*$h_341nUe",
	}

	// Adjust timeouts
	cluster.ConnectTimeout = 10 * time.Second
	cluster.Timeout = 10 * time.Second

	// Create a session
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("Failed to connect to the cluster: %v", err)
	}
	defer session.Close()

	// Execute a simple query to test the connection
	var clusterName string
	if err := session.Query(`SELECT cluster_name FROM system.local`).Scan(&clusterName); err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	fmt.Printf("Cluster Name: %s\n", clusterName)
}
