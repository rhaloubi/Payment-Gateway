package migrations

import (
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits/logger"
	model "github.com/rhaloubi/payment-gateway/tokenization-service/internal/models"
	"go.uber.org/zap"
)

func RunMigrations() error {
	db := inits.DB

	// Enable UUID extension (if not already enabled)
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		logger.Log.Error("failed to create uuid extension:", zap.Error(err))
	}

	// Auto migrate all models
	models := []interface{}{
		&model.CardBINInfo{},
		&model.CardVault{},
		&model.EncryptionKeyMetadata{},
		&model.TokenUsageLog{},
		&model.TokenizationRequest{},
	}

	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			logger.Log.Error("failed to migrate %T:", zap.Error(err))
		}
	}

	return nil
}

// RollbackMigrations rolls back all tokenization service migrations
func RollbackMigrations() error {
	db := inits.DB

	// Drop tables in reverse order
	models := []interface{}{
		&model.CardBINInfo{},
		&model.CardVault{},
		&model.EncryptionKeyMetadata{},
		&model.TokenUsageLog{},
		&model.TokenizationRequest{},
	}

	for _, m := range models {
		if err := db.Migrator().DropTable(m); err != nil {
			logger.Log.Error("failed to drop table %T:", zap.Error(err))
		}
	}

	return nil
}
