package main

import (
	"fmt"

	"github.com/rhaloubi/payment-gateway/auth-service/inits"
	"github.com/rhaloubi/payment-gateway/auth-service/inits/logger"
	model "github.com/rhaloubi/payment-gateway/auth-service/internal/models"
	"go.uber.org/zap"
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
		return fmt.Errorf("failed to create uuid extension: %w", err)
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
			return fmt.Errorf("failed to migrate %T: %w", m, err)
		}
	}

	// Seed default roles and permissions
	if err := seedDefaultRolesAndPermissions(); err != nil {
		return fmt.Errorf("failed to seed default data: %w", err)
	}

	return nil
}

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
			return fmt.Errorf("failed to create permission: %w", err)
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
			return fmt.Errorf("failed to create role: %w", err)
		}
	}

	// ============================================
	// STEP 3: ASSIGN PERMISSIONS TO ROLES
	// ============================================

	// ADMIN ROLE: Gets ALL permissions
	adminRole := roles[0]
	for _, perm := range permissions {
		if err := db.Model(&adminRole).Association("Permissions").Append(&perm); err != nil {
			return fmt.Errorf("failed to assign permission to admin: %w", err)
		}
	}

	// MANAGER ROLE: Can manage transactions and invoices, but NOT users or settings
	managerRole := roles[1]
	for _, perm := range permissions {
		// Give access to transactions and invoices
		if perm.Resource == "transactions" || perm.Resource == "invoices" {
			if err := db.Model(&managerRole).Association("Permissions").Append(&perm); err != nil {
				return fmt.Errorf("failed to assign permission to manager: %w", err)
			}
		}
		// Only READ access to settings and users (not update/create)
		if (perm.Resource == "settings" || perm.Resource == "users") && perm.Action == "read" {
			if err := db.Model(&managerRole).Association("Permissions").Append(&perm); err != nil {
				return fmt.Errorf("failed to assign permission to manager: %w", err)
			}
		}
	}

	// STAFF ROLE: Can only READ and CREATE transactions (no refunds/voids)
	staffRole := roles[2]
	for _, perm := range permissions {
		// Only basic transaction operations
		if perm.Resource == "transactions" && (perm.Action == "read" || perm.Action == "create") {
			if err := db.Model(&staffRole).Association("Permissions").Append(&perm); err != nil {
				return fmt.Errorf("failed to assign permission to staff: %w", err)
			}
		}
		// Can view invoices (but not create/update them)
		if perm.Resource == "invoices" && perm.Action == "read" {
			if err := db.Model(&staffRole).Association("Permissions").Append(&perm); err != nil {
				return fmt.Errorf("failed to assign permission to staff: %w", err)
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
			return fmt.Errorf("failed to drop table %T: %w", m, err)
		}
	}

	return nil
}
