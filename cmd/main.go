package main

// main.go
// @title           ChaikaReports API
// @version         1.0.4
// @description     API documentation for the ChaikaReports microservice.
// @termsOfService  http://swagger.io/terms/

// @contact.name    API Support
// @contact.url     http://www.swagger.io/support
// @contact.email   support@chaikareports.com

// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT

// @host            localhost:8080
// @BasePath        /api/v1

import (
	_ "ChaikaReports/cmd/docs"
	"ChaikaReports/internal/config"
	httpHandler "ChaikaReports/internal/handler/http"
	"ChaikaReports/internal/repository/cassandra"
	"ChaikaReports/internal/service"
	"context"
	"fmt"
	"github.com/go-kit/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yml"
	}

	// Initialize logger
	logger := log.NewLogfmtLogger(log.StdlibWriter{})
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	fmt.Printf("Loaded config: %+v\n", cfg)

	// Initialize Cassandra session
	session, err := cassandra.InitCassandra(logger, cfg.Cassandra.Keyspace, cfg.Cassandra.Hosts, cfg.Cassandra.User, cfg.Cassandra.Password, cfg.Cassandra.Timeout, cfg.Cassandra.RetryDelay, cfg.Cassandra.RetryAttempts)
	if err != nil {
		_ = logger.Log("error", "Failed to initialize Cassandra", "err", err)
		return
	}
	defer cassandra.CloseCassandra(session)

	// Initialize repository
	repo := cassandra.NewSalesRepository(session, logger)

	// Initialize service
	svc := service.NewSalesService(repo)

	// Initialize HTTP handler
	handler := httpHandler.NewHTTPHandler(svc, logger)

	// Start HTTP server
	// Create server with configurable port
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: handler,
	}

	// Graceful shutdown handling
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_ = logger.Log("msg", fmt.Sprintf("Starting server on %s", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			_ = logger.Log("error", "Failed to start server", "err", err)
			done <- os.Interrupt
		}
	}()

	<-done
	_ = logger.Log("msg", "Server stopping")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.Timeout*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		_ = logger.Log("error", "Server shutdown failed", "err", err)
	}

	_ = logger.Log("msg", "Server stopped")
}
