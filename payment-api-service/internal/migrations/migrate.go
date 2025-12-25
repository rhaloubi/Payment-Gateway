package migrations

import (
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits/logger"
	model "github.com/rhaloubi/payment-gateway/payment-api-service/internal/models"
	"go.uber.org/zap"
)

func RunPaymentApiMigrations() error {
	db := inits.DB

	// Enable UUID extension (if not already enabled)
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		logger.Log.Error("failed to create uuid extension:", zap.Error(err))
	}

	// Auto migrate all models
	models := []interface{}{
		&model.Payment{},
		&model.PaymentEvent{},
		&model.WebhookDelivery{},
		&model.PaymentIntent{}, // NEW
	}

	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			logger.Log.Error("failed to migrate %T:", zap.Error(err))
		}
	}

	// Create indexes for payment intents
	db.Exec("CREATE INDEX IF NOT EXISTS idx_payment_intents_merchant_id ON payment_intents(merchant_id);")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_payment_intents_status ON payment_intents(status);")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_payment_intents_expires_at ON payment_intents(expires_at);")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_payment_intents_order_id ON payment_intents(order_id);")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_intents_client_secret ON payment_intents(client_secret);")

	return nil
}

// RollbackPaymentApiMigrations rolls back all payment api service migrations
func RollbackPaymentApiMigrations() error {
	db := inits.DB

	// Drop tables in reverse order
	models := []interface{}{
		&model.WebhookDelivery{},
		&model.PaymentEvent{},
		&model.Payment{},
	}

	for _, m := range models {
		if err := db.Migrator().DropTable(m); err != nil {
			logger.Log.Error("failed to drop table %T:", zap.Error(err))
		}
	}

	return nil
}
