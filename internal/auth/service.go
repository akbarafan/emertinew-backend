package auth

import (
	"context"
	"errors"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/Alugara-Inovasi-Utama/emerti/middleware"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Login(username, password, ipAddress, userAgent string) (*LoginResponse, error)
	Logout(userID int, token, username, ipAddress, userAgent string) error
}

type service struct {
	repo         Repository
	auditService audit.Service
}

func NewService(repo Repository, auditService audit.Service) Service {
	return &service{
		repo:         repo,
		auditService: auditService,
	}
}

func (s *service) Login(username, password, ipAddress, userAgent string) (*LoginResponse, error) {
	// Find user
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		// Audit: login failed - user not found
		_ = s.auditService.Log(0, "auth", 0, "LOGIN_FAILED", nil, AuthAuditData{
			Username: username,
			Reason:   "user not found",
		}, ipAddress, userAgent)
		return nil, errors.New("invalid credentials")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		// Audit: login failed - wrong password
		_ = s.auditService.Log(user.ID, "auth", user.ID, "LOGIN_FAILED", nil, AuthAuditData{
			Username: username,
			Reason:   "invalid password",
		}, ipAddress, userAgent)
		return nil, errors.New("invalid credentials")
	}

	var companyID *uint
	var companyName string
	var companyCode string
	if user.Company != nil {
		companyID = &user.Company.ID
		companyName = user.Company.Name
		companyCode = user.Company.Code
	} else if user.CompanyID != nil {
		companyID = user.CompanyID
	}

	// Generate JWT token using middleware GenerateToken
	token, err := middleware.GenerateToken(int(user.ID), user.Username, "", user.Role, companyID, companyName)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}
	// Store session in Redis
	ctx := context.Background()
	if err := middleware.StoreSession(ctx, int(user.ID), token); err != nil {
		return nil, errors.New("failed to create session")
	}

	// Audit: login success
	_ = s.auditService.Log(user.ID, "auth", user.ID, "LOGIN", nil, AuthAuditData{
		Username: user.Username,
		Role:     user.Role,
	}, ipAddress, userAgent)

	return &LoginResponse{
		Token:       token,
		UserID:      user.ID,
		Username:    user.Username,
		Role:        user.Role,
		CompanyID:   companyID,
		CompanyName: companyName,
		CompanyCode: companyCode,
	}, nil
}

func (s *service) Logout(userID int, token, username, ipAddress, userAgent string) error {
	ctx := context.Background()
	err := middleware.InvalidateSession(ctx, userID, token)
	if err != nil {
		return err
	}

	// Audit: logout
	_ = s.auditService.Log(uint(userID), "auth", uint(userID), "LOGOUT", nil, AuthAuditData{
		Username: username,
	}, ipAddress, userAgent)

	return nil
}
