package repository

import (
	"encoding/json"
	"errors"
	"fmt"
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

// Cache keys for sessions
const (
	sessionCacheKeyByID    = "session:id:%s"
	sessionCacheKeyByToken = "session:token:%s"
	sessionCacheTTL        = 15 * time.Minute
)

// Create creates a new session
func (r *SessionRepository) Create(session *model.Session) error {
	err := inits.DB.Create(session).Error
	if err != nil {
		return err
	}

	// Cache the session
	r.cacheSession(session)

	return nil
}

// FindByID finds a session by ID (with Redis caching)
func (r *SessionRepository) FindByID(id uuid.UUID) (*model.Session, error) {
	// Try cache first
	cacheKey := fmt.Sprintf(sessionCacheKeyByID, id.String())
	cachedSession, err := inits.RDB.Get(inits.Ctx, cacheKey).Result()

	if err == nil && cachedSession != "" {
		var session model.Session
		if err = json.Unmarshal([]byte(cachedSession), &session); err == nil {
			return &session, nil
		}
	}

	// Get from database
	var session model.Session
	err = inits.DB.Where("id = ?", id).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}

	// Cache the session
	r.cacheSession(&session)

	return &session, nil
}

// FindByToken finds a session by JWT token hash (with Redis caching)
func (r *SessionRepository) FindByToken(tokenHash string) (*model.Session, error) {
	// Try cache first
	cacheKey := fmt.Sprintf(sessionCacheKeyByToken, tokenHash)
	cachedSession, err := inits.RDB.Get(inits.Ctx, cacheKey).Result()

	if err == nil && cachedSession != "" {
		var session model.Session
		if err = json.Unmarshal([]byte(cachedSession), &session); err == nil {
			return &session, nil
		}
	}

	// Get from database
	var session model.Session
	err = inits.DB.Where("jwt_token = ?", tokenHash).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}

	// Cache the session
	r.cacheSession(&session)

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
	err := inits.DB.Save(session).Error
	if err != nil {
		return err
	}

	// Update cache
	r.cacheSession(session)

	return nil
}

// RevokeSession revokes a session
func (r *SessionRepository) RevokeSession(id uuid.UUID) error {
	session, err := r.FindByID(id)
	if err != nil {
		return err
	}

	err = inits.DB.Model(&model.Session{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_revoked": true,
		}).Error

	if err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateSessionCache(id, session.JWTToken)

	return nil
}

// RevokeAllUserSessions revokes all sessions for a user
func (r *SessionRepository) RevokeAllUserSessions(userID uuid.UUID) error {
	// Get all user sessions first to invalidate cache
	sessions, _ := r.FindByUserID(userID)

	err := inits.DB.Model(&model.Session{}).
		Where("user_id = ? AND is_revoked = false", userID).
		Updates(map[string]interface{}{
			"is_revoked": true,
		}).Error

	if err != nil {
		return err
	}

	// Invalidate all session caches
	for _, session := range sessions {
		r.invalidateSessionCache(session.ID, session.JWTToken)
	}

	return nil
}

// DeleteExpiredSessions deletes all expired sessions
func (r *SessionRepository) DeleteExpiredSessions() error {
	return inits.DB.Where("expires_at < ?", time.Now()).Delete(&model.Session{}).Error
}

// IsSessionValid checks if a session is valid (with Redis caching)
func (r *SessionRepository) IsSessionValid(tokenHash string) (bool, error) {
	// Try cache first
	cacheKey := fmt.Sprintf(sessionCacheKeyByToken, tokenHash)
	cachedSession, err := inits.RDB.Get(inits.Ctx, cacheKey).Result()

	if err == nil && cachedSession != "" {
		var session model.Session
		if err = json.Unmarshal([]byte(cachedSession), &session); err == nil {
			return session.IsActive(), nil
		}
	}

	// Check database
	var count int64
	err = inits.DB.Model(&model.Session{}).
		Where("jwt_token = ? AND is_revoked = false AND expires_at > ?", tokenHash, time.Now()).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Helper: Cache session in Redis
func (r *SessionRepository) cacheSession(session *model.Session) {
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return
	}

	// Calculate TTL based on session expiry
	ttl := time.Until(session.ExpiresAt)
	if ttl > sessionCacheTTL {
		ttl = sessionCacheTTL
	}
	if ttl <= 0 {
		return // Don't cache expired sessions
	}

	// Cache by ID
	cacheKeyID := fmt.Sprintf(sessionCacheKeyByID, session.ID.String())
	inits.RDB.Set(inits.Ctx, cacheKeyID, sessionJSON, ttl)

	// Cache by token
	cacheKeyToken := fmt.Sprintf(sessionCacheKeyByToken, session.JWTToken)
	inits.RDB.Set(inits.Ctx, cacheKeyToken, sessionJSON, ttl)
}

// Helper: Invalidate session cache
func (r *SessionRepository) invalidateSessionCache(sessionID uuid.UUID, tokenHash string) {
	cacheKeyID := fmt.Sprintf(sessionCacheKeyByID, sessionID.String())
	cacheKeyToken := fmt.Sprintf(sessionCacheKeyByToken, tokenHash)

	inits.RDB.Del(inits.Ctx, cacheKeyID, cacheKeyToken)
}
