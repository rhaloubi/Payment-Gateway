package service

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/rhaloubi/payment-gateway/merchant-service/config"
)

// toNullString converts a string to sql.NullString
func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// toNullTime converts a time to sql.NullTime
func toNullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}

// getEnv gets environment variable with default
func getEnv(key, defaultValue string) string {
	if value := config.GetEnv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets environment variable as int with default
func getEnvInt(key string, defaultValue int) int {
	if value := config.GetEnv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
