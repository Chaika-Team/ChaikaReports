package main

// main.go
// @title           ChaikaReports API
// @version         1.0.7
// @description     API documentation for the ChaikaReports microservice.
// @termsOfService  http://swagger.io/terms/
//
// @contact.name    API Support
// @contact.url     http://www.swagger.io/support
// @contact.email   support@chaikareports.com
//
// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT
//
// @host            chaika-soft.ru
// @BasePath        /api/v1/report

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "ChaikaReports/cmd/docs"
	"ChaikaReports/internal/config"
	grpcHandler "ChaikaReports/internal/handler/grpc"
	httpHandler "ChaikaReports/internal/handler/http"
	"ChaikaReports/internal/repository/cassandra"
	"ChaikaReports/internal/service"

	"github.com/go-kit/log"
	"google.golang.org/grpc"
)

func main() {
	// ——— Load config ———
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yml"
	}
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	fmt.Printf("Loaded config: %+v\n", cfg)

	// ——— Logger ———
	logger := log.NewLogfmtLogger(log.StdlibWriter{})
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	// ——— Cassandra session ———
	session, err := cassandra.InitCassandra(
		logger,
		cfg.Cassandra.Keyspace,
		cfg.Cassandra.Hosts,
		cfg.Cassandra.User,
		cfg.Cassandra.Password,
		cfg.Cassandra.Timeout,
		cfg.Cassandra.RetryDelay,
		cfg.Cassandra.RetryAttempts,
	)
	if err != nil {
		_ = logger.Log("error", "Failed to initialize Cassandra", "err", err)
		return
	}
	defer cassandra.CloseCassandra(session)

	// ——— Wire up repo, service, handlers ———
	repo := cassandra.NewSalesRepository(session, logger)
	svc := service.NewSalesService(repo)
	httpSrvHandler := httpHandler.NewHTTPHandler(svc, logger)

	// ——— HTTP server ———
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.HTTPServer.Host, cfg.HTTPServer.Port),
		Handler: httpSrvHandler,
	}

	// ——— gRPC server ———
	grpcAddr := fmt.Sprintf("%s:%s", cfg.GRPCServer.Host, cfg.GRPCServer.Port)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		_ = logger.Log("error", "Failed to listen on gRPC address", "addr", grpcAddr, "err", err)
		return
	}
	grpcSrv := grpc.NewServer()
	router := grpcHandler.NewRouter(svc, logger)
	grpcHandler.RegisterGRPCServer(grpcSrv, router)

	// ——— Start both servers concurrently & handle graceful shutdown ———
	done := make(chan error, 2)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Forward OS signals into the error channel
	go func() {
		sig := <-sigChan
		done <- fmt.Errorf("received signal: %v", sig)
	}()

	go func() {
		_ = logger.Log("msg", "starting HTTP server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			done <- fmt.Errorf("http server error: %w", err)
		}
	}()

	go func() {
		_ = logger.Log("msg", "starting gRPC server", "addr", grpcAddr)
		if err := grpcSrv.Serve(lis); err != nil {
			done <- fmt.Errorf("grpc server error: %w", err)
		}
	}()

	// Wait for first error or signal
	err = <-done
	_ = logger.Log("msg", "shutting down servers", "reason", err)

	// Graceful HTTP shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTPServer.Timeout*time.Second)
	defer cancel()
	if shutdownErr := srv.Shutdown(ctx); shutdownErr != nil {
		_ = logger.Log("error", "HTTP graceful shutdown failed", "err", shutdownErr)
	}

	// Graceful gRPC shutdown
	grpcSrv.GracefulStop()
	_ = logger.Log("msg", "servers stopped")
}
