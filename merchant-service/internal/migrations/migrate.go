package migrations

import (
	"github.com/rhaloubi/payment-gateway/merchant-service/inits"
	"github.com/rhaloubi/payment-gateway/merchant-service/inits/logger"
	model "github.com/rhaloubi/payment-gateway/merchant-service/internal/models"
	"go.uber.org/zap"
)

func RunMerchantMigrations() error {
	db := inits.DB

	// Enable UUID extension (if not already enabled)
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		logger.Log.Error("failed to create uuid extension:", zap.Error(err))
	}

	// Auto migrate all models
	models := []interface{}{
		&model.Merchant{},
		&model.MerchantUser{},
		&model.MerchantInvitation{},
		&model.MerchantSettings{},
		&model.MerchantBusinessInfo{},
		&model.MerchantBranding{},
		&model.MerchantVerification{},
		&model.MerchantActivityLog{},
	}

	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			logger.Log.Error("failed to migrate %T:", zap.Error(err))
		}
	}

	return nil
}

// RollbackMerchantMigrations rolls back all merchant service migrations
func RollbackMerchantMigrations() error {
	db := inits.DB

	// Drop tables in reverse order
	models := []interface{}{
		&model.MerchantActivityLog{},
		&model.MerchantVerification{},
		&model.MerchantBranding{},
		&model.MerchantBusinessInfo{},
		&model.MerchantSettings{},
		&model.MerchantInvitation{},
		&model.MerchantUser{},
		&model.Merchant{},
	}

	for _, m := range models {
		if err := db.Migrator().DropTable(m); err != nil {
			logger.Log.Error("failed to drop table %T:", zap.Error(err))
		}
	}

	return nil
}
