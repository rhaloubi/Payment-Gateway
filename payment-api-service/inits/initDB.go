package inits

import (
	"time"

	"github.com/rhaloubi/payment-gateway/payment-api-service/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	dsn := config.GetEnv("DATABASE_DSN")
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	sqlDB, err := DB.DB()
	if err != nil {
		panic("failed to get database instance")
	}

	sqlDB.SetMaxOpenConns(10)                  // max concurrent connections
	sqlDB.SetMaxIdleConns(5)                   // keep 5 idle for reuse
	sqlDB.SetConnMaxLifetime(time.Hour)        // refresh connections every hour
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // idle connections older than 10min are closed

}
