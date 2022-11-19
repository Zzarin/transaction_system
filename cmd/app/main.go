package main

import (
	"context"
	"github.com/Zzarin/transaction_system/internal"
	httpApi "github.com/Zzarin/transaction_system/internal/http"
	"github.com/Zzarin/transaction_system/internal/postgres"
	"github.com/Zzarin/transaction_system/internal/rabbitMQ"
	"github.com/jessevdk/go-flags"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	var cfg Config
	parser := flags.NewParser(&cfg, flags.Default)

	_, err := parser.Parse()
	if err != nil {
		log.Fatal("Error parse env variables", err)
	}

	logger := InitializeLoger()
	defer func() {
		err := logger.Sync()
		if err != nil {
			logger.Error("erase logs from buffer", zap.Error(err))
		}
	}()

	db, err := NewDb(cfg.DbDsn, cfg.DbConnMaxLifetime, cfg.DbMaxOpenConns, cfg.DbMaxIdleConns)
	if err != nil {
		logger.Fatal("get DB instance", zap.Error(err))
	}
	defer func() {
		err := db.Close()
		if err != nil {
			logger.Error("closing db connection", zap.Error(err))
		}
	}()
	logger.Info("Successful db connection")

	accountStorage := postgres.NewAccountSource(db)
	brokerConn, err := rabbitMQ.GetNewDistributor(cfg.AmqpConnectionURL)
	if err != nil {
		logger.Fatal("rabbitMQ connection", zap.Error(err))
	}
	defer func() {
		err := brokerConn.Conn.Close()
		if err != nil {
			logger.Error("closing rabbitMQ connection", zap.Error(err))
		}
	}()
	logger.Info("Broker connected")

	account := internal.NewAccountRepo(accountStorage, brokerConn, logger)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	handler := httpApi.InitEndpoints(ctx, account, logger)
	appServer := httpApi.NewServer(cfg.HostPort, handler)
	logger.Info("Starting httpServer...", zap.String("port", cfg.HostPort))

	err = appServer.Start()
	if err != nil {
		cancel()
		logger.Fatal("Couldn't start server on port", zap.Error(err))
	}

	go func() {
		<-ctx.Done()
		logger.Info("Shutdown signal received")
		err := appServer.Stop(ctx)
		if err != nil {
			logger.Error("stop the server", zap.Error(err))
		}
	}()
}
