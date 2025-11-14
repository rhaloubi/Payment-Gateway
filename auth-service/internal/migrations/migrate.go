package main

import (
	"github.com/rhaloubi/payment-gateway/auth-service/inits"
	"github.com/rhaloubi/payment-gateway/auth-service/inits/logger"
	model "github.com/rhaloubi/payment-gateway/auth-service/internal/models"
	"go.uber.org/zap"
	//"gorm.io/gorm"
)

func init() {
	inits.InitDotEnv()
	logger.Init()
	inits.InitDB()
}
func main() {
	// Run migrations
	if err := RunAuthMigrations(); err != nil {
		logger.Log.Error("Migration failed", zap.Error(err))
	}

	logger.Log.Info("✅ Migrations completed successfully!")
}

func RunAuthMigrations() error {
	// Enable UUID extension
	db := inits.DB

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		logger.Log.Error("failed to create uuid extension:", zap.Error(err))
	}

	// Auto migrate all models
	models := []interface{}{
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.UserRole{},
		&model.RolePermission{},
		&model.Session{},
		&model.APIKey{},
	}

	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			logger.Log.Error("failed to migrate:", zap.Error(err))
		}
	}

	/*if err := ensureUserRoleColumns(db); err != nil {
		return fmt.Errorf("failed to ensure UserRole merchant_id column: %w", err)
	}
	*/

	// Seed default roles and permissions
	if err := seedDefaultRolesAndPermissions(); err != nil {
		logger.Log.Error("failed to seed default data:", zap.Error(err))
	}

	return nil
}

/* ensureUserRoleColumns ensures the user_roles table has all required columns
func ensureUserRoleColumns(db *gorm.DB) error {
	// Check and add merchant_id column
	if err := ensureColumnExists(db, "user_roles", "merchant_id", "UUID NOT NULL"); err != nil {
		return fmt.Errorf("failed to ensure merchant_id column: %w", err)
	}

	// Check and add assigned_by column
	if err := ensureColumnExists(db, "user_roles", "assigned_by", "UUID"); err != nil {
		return fmt.Errorf("failed to ensure assigned_by column: %w", err)
	}

	// Check and add assigned_at column
	if err := ensureColumnExists(db, "user_roles", "assigned_at", "TIMESTAMP NOT NULL DEFAULT NOW()"); err != nil {
		return fmt.Errorf("failed to ensure assigned_at column: %w", err)
	}

	// Ensure index on merchant_id
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_user_roles_merchant_id ON user_roles(merchant_id)`).Error; err != nil {
		return fmt.Errorf("failed to create index on merchant_id: %w", err)
	}

	return nil
}

// ensureColumnExists checks if a column exists and adds it if it doesn't
func ensureColumnExists(db *gorm.DB, tableName, columnName, columnDefinition string) error {
	var columnExists bool
	err := db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = ?
			AND column_name = ?
		)
	`, tableName, columnName).Scan(&columnExists).Error

	if err != nil {
		return fmt.Errorf("failed to check if %s column exists: %w", columnName, err)
	}

	if !columnExists {
		// Add the column
		alterSQL := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, columnDefinition)
		if err := db.Exec(alterSQL).Error; err != nil {
			return fmt.Errorf("failed to add %s column: %w", columnName, err)
		}
	}

	return nil
}
*/

func seedDefaultRolesAndPermissions() error {
	db := inits.DB

	// Check if already seeded
	var count int64
	db.Model(&model.Role{}).Count(&count)
	if count > 0 {
		return nil // Already seeded
	}

	// ============================================
	// STEP 1: CREATE PERMISSIONS
	// ============================================
	// Permissions define individual actions on resources

	permissions := []model.Permission{
		// Transaction permissions
		{Resource: "transactions", Action: "read", Description: "View transaction details"},
		{Resource: "transactions", Action: "create", Description: "Create new transactions (auth/sale)"},
		{Resource: "transactions", Action: "refund", Description: "Refund transactions"},
		{Resource: "transactions", Action: "void", Description: "Void/cancel transactions"},

		// Invoice permissions
		{Resource: "invoices", Action: "read", Description: "View invoices"},
		{Resource: "invoices", Action: "create", Description: "Create invoices"},
		{Resource: "invoices", Action: "update", Description: "Update invoices"},

		// API Key permissions
		{Resource: "api_keys", Action: "read", Description: "View API keys"},
		{Resource: "api_keys", Action: "create", Description: "Create API keys"},
		{Resource: "api_keys", Action: "delete", Description: "Delete API keys"},

		// User management permissions
		{Resource: "users", Action: "read", Description: "View team members"},
		{Resource: "users", Action: "create", Description: "Invite team members"},
		{Resource: "users", Action: "update", Description: "Update team member roles"},

		// Settings permissions
		{Resource: "settings", Action: "read", Description: "View merchant settings"},
		{Resource: "settings", Action: "update", Description: "Update merchant settings"},
	}

	// Create permissions
	for i := range permissions {
		if err := db.Create(&permissions[i]).Error; err != nil {
			logger.Log.Error("failed to create permission:", zap.Error(err))
			//	return fmt.Errorf("failed to create permission: %w", err)
		}
	}

	// ============================================
	// STEP 2: CREATE ROLES
	// ============================================
	// Roles are collections of permissions

	roles := []model.Role{
		{
			Name:        "Admin",
			Description: "Full access to everything - can manage payments, users, settings",
		},
		{
			Name:        "Manager",
			Description: "Can manage payments and invoices, but cannot change settings or manage users",
		},
		{
			Name:        "Staff",
			Description: "Can only view and create transactions - no refunds or management access",
		},
	}

	// Create roles
	for i := range roles {
		if err := db.Create(&roles[i]).Error; err != nil {
			logger.Log.Error("failed to create role:", zap.Error(err))
		}
	}

	// ============================================
	// STEP 3: ASSIGN PERMISSIONS TO ROLES
	// ============================================

	// ADMIN ROLE: Gets ALL permissions
	adminRole := roles[0]
	for _, perm := range permissions {
		if err := db.Model(&adminRole).Association("Permissions").Append(&perm); err != nil {
			logger.Log.Error("failed to assign permission to admin:", zap.Error(err))
		}
	}

	// MANAGER ROLE: Can manage transactions and invoices, but NOT users or settings
	managerRole := roles[1]
	for _, perm := range permissions {
		// Give access to transactions and invoices
		if perm.Resource == "transactions" || perm.Resource == "invoices" {
			if err := db.Model(&managerRole).Association("Permissions").Append(&perm); err != nil {
				logger.Log.Error("failed to assign permission to manager:", zap.Error(err))
			}
		}
		// Only READ access to settings and users (not update/create)
		if (perm.Resource == "settings" || perm.Resource == "users") && perm.Action == "read" {
			if err := db.Model(&managerRole).Association("Permissions").Append(&perm); err != nil {
				logger.Log.Error("failed to assign permission to manager:", zap.Error(err))
			}
		}
	}

	// STAFF ROLE: Can only READ and CREATE transactions (no refunds/voids)
	staffRole := roles[2]
	for _, perm := range permissions {
		// Only basic transaction operations
		if perm.Resource == "transactions" && (perm.Action == "read" || perm.Action == "create") {
			if err := db.Model(&staffRole).Association("Permissions").Append(&perm); err != nil {
				logger.Log.Error("failed to assign permission to staff:", zap.Error(err))
			}
		}
		// Can view invoices (but not create/update them)
		if perm.Resource == "invoices" && perm.Action == "read" {
			if err := db.Model(&staffRole).Association("Permissions").Append(&perm); err != nil {
				logger.Log.Error("failed to assign permission to staff:", zap.Error(err))
			}
		}
	}

	return nil
}

func RollbackAuthMigrations() error {
	db := inits.DB
	// Drop tables in reverse order
	models := []interface{}{
		&model.APIKey{},
		&model.Session{},
		&model.RolePermission{},
		&model.UserRole{},
		&model.Permission{},
		&model.Role{},
		&model.User{},
	}

	for _, m := range models {
		if err := db.Migrator().DropTable(m); err != nil {
			logger.Log.Error("failed to drop table:", zap.Error(err))
		}
	}

	return nil
}
