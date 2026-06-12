package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"new-forstitch-site/backend/internal/models"
)

const AdminSessionTTL = 12 * time.Hour
const CustomerSessionTTL = 30 * 24 * time.Hour

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

func (s *Service) CustomerLogin(req models.LoginRequest) (models.CustomerSessionResponse, models.CustomerSession, time.Time, error) {
	email := normalizeEmail(req.Username)
	if email == "" || req.Password == "" {
		return models.CustomerSessionResponse{}, models.CustomerSession{}, time.Time{}, models.Validation("credentials_required", "email and password are required")
	}

	user, err := s.repo.CustomerByEmail(email)
	if err != nil {
		return models.CustomerSessionResponse{}, models.CustomerSession{}, time.Time{}, models.Unauthorized("invalid_credentials", "invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return models.CustomerSessionResponse{}, models.CustomerSession{}, time.Time{}, models.Unauthorized("invalid_credentials", "invalid email or password")
	}

	sessionID, err := randomToken()
	if err != nil {
		return models.CustomerSessionResponse{}, models.CustomerSession{}, time.Time{}, err
	}
	session := models.CustomerSession{ID: sessionID, UserID: user.ID, Email: user.Email, Name: user.Name}
	expiresAt := time.Now().Add(CustomerSessionTTL)
	if err := s.repo.CreateCustomerSession(session, expiresAt); err != nil {
		return models.CustomerSessionResponse{}, models.CustomerSession{}, time.Time{}, err
	}
	_ = s.repo.DeleteExpiredCustomerSessions(time.Now())

	return models.CustomerSessionResponse{Authenticated: true, Email: user.Email, Name: user.Name}, session, expiresAt, nil
}

func (s *Service) StartCustomerRegistration(req models.CustomerRegistrationStartRequest) (models.CustomerRegistrationStartResponse, error) {
	email := normalizeEmail(req.Email)
	name := strings.TrimSpace(req.Name)
	if _, err := mail.ParseAddress(email); err != nil {
		return models.CustomerRegistrationStartResponse{}, models.Validation("email_invalid", "valid email is required")
	}
	if len(req.Password) < 6 {
		return models.CustomerRegistrationStartResponse{}, models.Validation("password_short", "password must be at least 6 characters")
	}
	if _, err := s.repo.CustomerByEmail(email); err == nil {
		return models.CustomerRegistrationStartResponse{}, models.Conflict("customer_exists", "customer already exists")
	} else if !errors.Is(err, models.ErrNotFound) {
		return models.CustomerRegistrationStartResponse{}, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.CustomerRegistrationStartResponse{}, err
	}
	code, err := randomCode()
	if err != nil {
		return models.CustomerRegistrationStartResponse{}, err
	}
	if err := s.repo.SaveCustomerRegistrationCode(email, name, string(passwordHash), hashString(code), time.Now().Add(registrationCodeTTL)); err != nil {
		return models.CustomerRegistrationStartResponse{}, err
	}
	if err := s.sendRegistrationCode(email, code); err != nil {
		return models.CustomerRegistrationStartResponse{}, err
	}
	return models.CustomerRegistrationStartResponse{
		Email:   email,
		Message: "Код подтверждения отправлен на email.",
	}, nil
}

func (s *Service) VerifyCustomerRegistration(req models.CustomerRegistrationVerifyRequest) (models.CustomerSessionResponse, models.CustomerSession, time.Time, error) {
	email := normalizeEmail(req.Email)
	code := strings.TrimSpace(req.Code)
	if email == "" || code == "" {
		return models.CustomerSessionResponse{}, models.CustomerSession{}, time.Time{}, models.Validation("code_required", "email and code are required")
	}

	user, err := s.repo.CustomerByRegistrationCode(email, hashString(code), time.Now())
	if err != nil {
		return models.CustomerSessionResponse{}, models.CustomerSession{}, time.Time{}, models.Unauthorized("code_invalid", "registration code is invalid")
	}
	_ = s.repo.DeleteCustomerRegistrationCode(email)

	sessionID, err := randomToken()
	if err != nil {
		return models.CustomerSessionResponse{}, models.CustomerSession{}, time.Time{}, err
	}
	session := models.CustomerSession{ID: sessionID, UserID: user.ID, Email: user.Email, Name: user.Name}
	expiresAt := time.Now().Add(CustomerSessionTTL)
	if err := s.repo.CreateCustomerSession(session, expiresAt); err != nil {
		return models.CustomerSessionResponse{}, models.CustomerSession{}, time.Time{}, err
	}
	_ = s.repo.DeleteExpiredCustomerSessions(time.Now())
	return models.CustomerSessionResponse{Authenticated: true, Email: user.Email, Name: user.Name}, session, expiresAt, nil
}

func (s *Service) StartCustomerPasswordReset(req models.CustomerPasswordResetStartRequest) (models.CustomerRegistrationStartResponse, error) {
	email := normalizeEmail(req.Email)
	if _, err := mail.ParseAddress(email); err != nil {
		return models.CustomerRegistrationStartResponse{}, models.Validation("email_invalid", "valid email is required")
	}
	if _, err := s.repo.CustomerByEmail(email); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return models.CustomerRegistrationStartResponse{Email: email, Message: "Если email зарегистрирован, код восстановления отправлен."}, nil
		}
		return models.CustomerRegistrationStartResponse{}, err
	}
	code, err := randomCode()
	if err != nil {
		return models.CustomerRegistrationStartResponse{}, err
	}
	if err := s.repo.SaveCustomerPasswordResetCode(email, hashString(code), time.Now().Add(registrationCodeTTL)); err != nil {
		return models.CustomerRegistrationStartResponse{}, err
	}
	if err := s.sendPasswordResetCode(email, code); err != nil {
		return models.CustomerRegistrationStartResponse{}, err
	}
	return models.CustomerRegistrationStartResponse{Email: email, Message: "Код восстановления отправлен на email."}, nil
}

func (s *Service) VerifyCustomerPasswordReset(req models.CustomerPasswordResetVerifyRequest) (models.CustomerSessionResponse, models.CustomerSession, time.Time, error) {
	email := normalizeEmail(req.Email)
	code := strings.TrimSpace(req.Code)
	if email == "" || code == "" {
		return models.CustomerSessionResponse{}, models.CustomerSession{}, time.Time{}, models.Validation("code_required", "email and code are required")
	}
	if len(req.NewPassword) < 6 {
		return models.CustomerSessionResponse{}, models.CustomerSession{}, time.Time{}, models.Validation("password_short", "password must be at least 6 characters")
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return models.CustomerSessionResponse{}, models.CustomerSession{}, time.Time{}, err
	}
	user, err := s.repo.UpdateCustomerPasswordByResetCode(email, hashString(code), string(passwordHash), time.Now())
	if err != nil {
		return models.CustomerSessionResponse{}, models.CustomerSession{}, time.Time{}, models.Unauthorized("code_invalid", "password reset code is invalid")
	}
	_ = s.repo.DeleteCustomerPasswordResetCode(email)

	sessionID, err := randomToken()
	if err != nil {
		return models.CustomerSessionResponse{}, models.CustomerSession{}, time.Time{}, err
	}
	session := models.CustomerSession{ID: sessionID, UserID: user.ID, Email: user.Email, Name: user.Name}
	expiresAt := time.Now().Add(CustomerSessionTTL)
	if err := s.repo.CreateCustomerSession(session, expiresAt); err != nil {
		return models.CustomerSessionResponse{}, models.CustomerSession{}, time.Time{}, err
	}
	_ = s.repo.DeleteExpiredCustomerSessions(time.Now())
	return models.CustomerSessionResponse{Authenticated: true, Email: user.Email, Name: user.Name}, session, expiresAt, nil
}

func (s *Service) CustomerSession(sessionID string) (models.CustomerSession, error) {
	if strings.TrimSpace(sessionID) == "" {
		return models.CustomerSession{}, models.Unauthorized("session_required", "customer session is required")
	}
	return s.repo.CustomerSession(sessionID, time.Now())
}

func (s *Service) CustomerLogout(sessionID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return nil
	}
	return s.repo.DeleteCustomerSession(sessionID)
}

func (s *Service) CustomerOrders(customerID int64) ([]models.Order, error) {
	return s.repo.CustomerOrders(customerID)
}

func (s *Service) CustomerOrder(orderID string, customerID int64) (models.Order, error) {
	if strings.TrimSpace(orderID) == "" {
		return models.Order{}, models.Validation("order_id_required", "order id is required")
	}
	return s.repo.OrderForCustomer(orderID, customerID)
}

func randomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func hashString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
