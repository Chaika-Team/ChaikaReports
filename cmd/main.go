package main

import (
	"ChaikaReports/internal/config"
	httpHandler "ChaikaReports/internal/handler/http"
	"ChaikaReports/internal/repository/cassandra"
	"ChaikaReports/internal/service"
	"net/http"

	"github.com/go-kit/log"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig("config.yml")

	// Initialize logger
	logger := log.NewLogfmtLogger(log.StdlibWriter{})
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	// Initialize Cassandra session
	session, err := cassandra.InitCassandra(logger, cfg.Cassandra.Keyspace, cfg.Cassandra.Hosts, cfg.Cassandra.User, cfg.Cassandra.Password)
	if err != nil {
		logger.Log("error", "Failed to initialize Cassandra", "err", err)
		return
	}
	defer cassandra.CloseCassandra(session)

	// Initialize repository
	repo := cassandra.NewSalesRepository(session, logger)

	// Initialize service
	svc := service.NewSalesService(repo)

	// Initialize HTTP handler
	handler := httpHandler.NewHTTPHandler(svc)

	// Start HTTP server
	logger.Log("msg", "Starting server on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		logger.Log("error", "Failed to start server", "err", err)
	}
}
