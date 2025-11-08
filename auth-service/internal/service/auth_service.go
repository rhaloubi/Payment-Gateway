package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/auth-service/inits"
	"github.com/rhaloubi/payment-gateway/auth-service/inits/jwt"
	model "github.com/rhaloubi/payment-gateway/auth-service/internal/models"
	"github.com/rhaloubi/payment-gateway/auth-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo     *repository.UserRepository
	sessionRepo  *repository.SessionRepository
	jwtUtil      *jwt.JWTUtil
	emailService *inits.EmailService
}

func NewAuthService() *AuthService {
	return &AuthService{
		userRepo:     repository.NewUserRepository(),
		sessionRepo:  repository.NewSessionRepository(),
		jwtUtil:      jwt.NewJWTUtil(),
		emailService: inits.NewEmailService(),
	}
}

type RegisterRequest struct {
	Name     string
	Email    string
	Password string
}

type LoginRequest struct {
	Email     string
	Password  string
	IPAddress string
	UserAgent string
}

type LoginResponse struct {
	User         *model.User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 // seconds
}

// Register creates a new user account
func (s *AuthService) Register(req *RegisterRequest) (*model.User, error) {

	if err := s.validateRegistration(req); err != nil {
		return nil, err
	}

	// Check if email already exists
	exists, err := s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create user
	user := &model.User{
		Name:          req.Name,
		Email:         req.Email,
		PasswordHash:  string(hashedPassword),
		Status:        model.UserStatusPendingVerification,
		EmailVerified: false,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and creates a session
func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check if account is locked
	if user.IsLocked() {
		return nil, errors.New("account is locked due to too many failed login attempts")
	}

	// Check if account is suspended
	if user.Status == model.UserStatusSuspended {
		return nil, errors.New("account is suspended")
	}

	// Verify password
	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		// Increment failed login attempts
		s.userRepo.IncrementFailedLoginAttempts(user.ID)

		// Lock account after 5 failed attempts
		if user.FailedLoginAttempts >= 4 {
			lockUntil := time.Now().Add(30 * time.Minute)
			s.userRepo.LockAccount(user.ID, lockUntil)
			return nil, errors.New("account locked due to too many failed login attempts")
		}

		return nil, errors.New("invalid email or password")
	}

	// Generate JWT tokens
	accessToken, err := s.jwtUtil.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	refreshToken, err := s.jwtUtil.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	// Create session
	tokenHash := s.jwtUtil.HashToken(accessToken)
	session := &model.Session{
		UserID:    user.ID,
		JWTToken:  tokenHash,
		IPAddress: toNullString(req.IPAddress),
		UserAgent: toNullString(req.UserAgent),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hours
		IsRevoked: false,
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return nil, errors.New("failed to create session")
	}

	// Update last login
	s.userRepo.UpdateLastLogin(user.ID, req.IPAddress)

	return &LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    86400, // 24 hours in seconds
	}, nil
}

// Logout revokes a user's session
func (s *AuthService) Logout(accessToken string) error {
	tokenHash := s.jwtUtil.HashToken(accessToken)

	session, err := s.sessionRepo.FindByToken(tokenHash)
	if err != nil {
		return errors.New("session not found")
	}

	return s.sessionRepo.RevokeSession(session.ID)
}

// LogoutAll revokes all sessions for a user
func (s *AuthService) LogoutAll(userID uuid.UUID) error {
	return s.sessionRepo.RevokeAllUserSessions(userID)
}

// ValidateToken validates an access token and returns the user
func (s *AuthService) ValidateToken(accessToken string) (*model.User, error) {
	// Parse and validate JWT
	claims, err := s.jwtUtil.ValidateAccessToken(accessToken)
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	// Check if session exists and is valid
	tokenHash := s.jwtUtil.HashToken(accessToken)
	isValid, err := s.sessionRepo.IsSessionValid(tokenHash)
	if err != nil || !isValid {
		return nil, errors.New("session not found or revoked")
	}

	// Get user
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, errors.New("invalid user ID in token")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Check if user is active
	if user.Status != model.UserStatusActive {
		return nil, errors.New("user account is not active")
	}

	return user, nil
}

// RefreshToken generates a new access token using a refresh token
func (s *AuthService) RefreshToken(refreshToken string) (*LoginResponse, error) {
	// Validate refresh token
	claims, err := s.jwtUtil.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	// Get user
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, errors.New("invalid user ID in token")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Generate new tokens
	newAccessToken, err := s.jwtUtil.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	newRefreshToken, err := s.jwtUtil.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	return &LoginResponse{
		User:         user,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    86400,
	}, nil
}

// VerifyEmail marks a user's email as verified
func (s *AuthService) VerifyEmail(userID uuid.UUID) error {
	return s.userRepo.VerifyEmail(userID)
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(userID uuid.UUID, oldPassword, newPassword string) error {
	// Get user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	// Validate new password
	if len(newPassword) < 8 {
		return errors.New("new password must be at least 8 characters")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// Update password
	user.PasswordHash = string(hashedPassword)
	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	// Revoke all sessions (force re-login)
	s.sessionRepo.RevokeAllUserSessions(userID)

	return nil
}

// GetUserSessions gets all active sessions for a user
func (s *AuthService) GetUserSessions(userID uuid.UUID) ([]model.Session, error) {
	return s.sessionRepo.FindByUserID(userID)
}

// validateRegistration validates registration input
func (s *AuthService) validateRegistration(req *RegisterRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}
	if req.Email == "" {
		return errors.New("email is required")
	}
	if req.Password == "" {
		return errors.New("password is required")
	}
	if len(req.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	// TODO: Add email format validation
	return nil
}
