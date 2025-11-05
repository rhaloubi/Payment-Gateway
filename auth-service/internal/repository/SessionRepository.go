package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/auth-service/inits"
	model "github.com/rhaloubi/payment-gateway/auth-service/internal/models"
	"gorm.io/gorm"
)

type SessionRepository struct{}

// NewSessionRepository creates a new session repository
func NewSessionRepository() *SessionRepository {
	return &SessionRepository{}
}

// Create creates a new session
func (r *SessionRepository) Create(session *model.Session) error {
	return inits.DB.Create(session).Error
}

// FindByID finds a session by ID
func (r *SessionRepository) FindByID(id uuid.UUID) (*model.Session, error) {
	var session model.Session
	err := inits.DB.Where("id = ?", id).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}
	return &session, nil
}

// FindByToken finds a session by JWT token hash
func (r *SessionRepository) FindByToken(tokenHash string) (*model.Session, error) {
	var session model.Session
	err := inits.DB.Where("jwt_token = ?", tokenHash).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}
	return &session, nil
}

// FindByUserID finds all sessions for a user
func (r *SessionRepository) FindByUserID(userID uuid.UUID) ([]model.Session, error) {
	var sessions []model.Session
	err := inits.DB.Where("user_id = ? AND is_revoked = false AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").
		Find(&sessions).Error
	return sessions, err
}

// Update updates a session
func (r *SessionRepository) Update(session *model.Session) error {
	return inits.DB.Save(session).Error
}

// RevokeSession revokes a session
func (r *SessionRepository) RevokeSession(id uuid.UUID) error {
	return inits.DB.Model(&model.Session{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_revoked": true,
		}).Error
}

// RevokeAllUserSessions revokes all sessions for a user
func (r *SessionRepository) RevokeAllUserSessions(userID uuid.UUID) error {
	return inits.DB.Model(&model.Session{}).
		Where("user_id = ? AND is_revoked = false", userID).
		Updates(map[string]interface{}{
			"is_revoked": true,
		}).Error
}

// DeleteExpiredSessions deletes all expired sessions
func (r *SessionRepository) DeleteExpiredSessions() error {
	return inits.DB.Where("expires_at < ?", time.Now()).Delete(&model.Session{}).Error
}

// IsSessionValid checks if a session is valid (not revoked and not expired)
func (r *SessionRepository) IsSessionValid(tokenHash string) (bool, error) {
	var count int64
	err := inits.DB.Model(&model.Session{}).
		Where("jwt_token = ? AND is_revoked = false AND expires_at > ?", tokenHash, time.Now()).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}
