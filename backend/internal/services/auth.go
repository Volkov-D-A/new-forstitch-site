package services

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"new-forstitch-site/backend/internal/models"
)

const AdminSessionTTL = 12 * time.Hour

func (s *Service) EnsureAdminUser(username string, password string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return models.Validation("username_required", "username is required")
	}
	if password == "" {
		return models.Validation("password_required", "password is required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.EnsureAdminUser(username, string(hash))
}

func (s *Service) Login(req models.LoginRequest) (models.LoginResponse, models.AdminSession, time.Time, error) {
	username := strings.TrimSpace(req.Username)
	if username == "" || req.Password == "" {
		return models.LoginResponse{}, models.AdminSession{}, time.Time{}, models.Validation("credentials_required", "username and password are required")
	}

	user, err := s.repo.AdminUserByUsername(username)
	if err != nil {
		return models.LoginResponse{}, models.AdminSession{}, time.Time{}, models.Unauthorized("invalid_credentials", "invalid username or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return models.LoginResponse{}, models.AdminSession{}, time.Time{}, models.Unauthorized("invalid_credentials", "invalid username or password")
	}

	sessionID, err := randomToken()
	if err != nil {
		return models.LoginResponse{}, models.AdminSession{}, time.Time{}, err
	}
	csrfToken, err := randomToken()
	if err != nil {
		return models.LoginResponse{}, models.AdminSession{}, time.Time{}, err
	}

	session := models.AdminSession{
		ID:        sessionID,
		UserID:    user.ID,
		Username:  user.Username,
		CSRFToken: csrfToken,
	}
	expiresAt := time.Now().Add(AdminSessionTTL)
	if err := s.repo.CreateAdminSession(session, expiresAt); err != nil {
		return models.LoginResponse{}, models.AdminSession{}, time.Time{}, err
	}
	_ = s.repo.DeleteExpiredAdminSessions(time.Now())

	return models.LoginResponse{Username: user.Username, CSRFToken: csrfToken}, session, expiresAt, nil
}

func (s *Service) Session(sessionID string) (models.AdminSession, error) {
	if strings.TrimSpace(sessionID) == "" {
		return models.AdminSession{}, models.Unauthorized("session_required", "admin session is required")
	}
	return s.repo.AdminSession(sessionID, time.Now())
}

func (s *Service) Logout(sessionID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return nil
	}
	return s.repo.DeleteAdminSession(sessionID)
}

func (s *Service) CheckCSRF(session models.AdminSession, csrfToken string) error {
	if csrfToken == "" || csrfToken != session.CSRFToken {
		return models.Unauthorized("csrf_invalid", "csrf token is invalid")
	}
	return nil
}

func randomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
