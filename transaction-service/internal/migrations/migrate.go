package main

import (
	"github.com/rhaloubi/payment-gateway/transaction-service/inits"
	"github.com/rhaloubi/payment-gateway/transaction-service/inits/logger"
	model "github.com/rhaloubi/payment-gateway/transaction-service/internal/models"
	"go.uber.org/zap"
)

func init() {
	inits.InitDotEnv()
	logger.Init()
	inits.InitDB()
}

func main() {
	// Run migrations
	if err := RunMigrations(); err != nil {
		logger.Log.Error("Migration failed", zap.Error(err))
	}

	logger.Log.Info("✅ Migrations completed successfully!")
}

func RunMigrations() error {
	db := inits.DB

	// Enable UUID extension (if not already enabled)
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		logger.Log.Error("failed to create uuid extension:", zap.Error(err))
	}

	// Auto migrate all models
	models := []interface{}{
		&model.Transaction{},
		&model.ChargebackEvent{},
		&model.ExchangeRate{},
		&model.TransactionEvent{},
		&model.Chargeback{},
		&model.SettlementBatch{},
		&model.IssuerResponse{},
	}

	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			logger.Log.Error("failed to migrate %T:", zap.Error(err))
		}
	}

	return nil
}

func RollbackMigrations() error {
	db := inits.DB

	// Drop tables in reverse order
	models := []interface{}{
		&model.Transaction{},
		&model.ChargebackEvent{},
		&model.ExchangeRate{},
		&model.TransactionEvent{},
		&model.Chargeback{},
		&model.SettlementBatch{},
		&model.IssuerResponse{},
	}

	for _, m := range models {
		if err := db.Migrator().DropTable(m); err != nil {
			logger.Log.Error("failed to drop table %T:", zap.Error(err))
		}
	}

	return nil
}
