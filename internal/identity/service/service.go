package service

import (
	"canbankx/internal/identity/domain"
	"canbankx/internal/identity/ports"
	"canbankx/internal/pkg/auth"
	apierrors "canbankx/internal/pkg/errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type IdentityService interface {
	RegisterUser(req RegisterRequest) (*domain.User, *apierrors.APIError)
	Login(req LoginRequest) (*LoginResponse, *apierrors.APIError)
	VerifyMFA(req VerifyMFARequest) (*TokenResponse, *apierrors.APIError)
}

type RegisterRequest struct {
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	SIN       string `json:"sin" binding:"required,len=9"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Message  string `json:"message"`
	MFAToken string `json:"mfaToken"`
}

type VerifyMFARequest struct {
	MFAToken string `json:"mfaToken" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

type TokenResponse struct {
	AccessToken string `json:"accessToken"`
}

type identityService struct {
	userRepo    ports.UserRepository
	sessionRepo ports.SessionRepository
}

func NewIdentityService(userRepo ports.UserRepository, sessionRepo ports.SessionRepository) IdentityService {
	return &identityService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

func (s *identityService) RegisterUser(req RegisterRequest) (*domain.User, *apierrors.APIError) {

	existingUser, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, apierrors.ErrInternalServer
	}
	if existingUser != nil {
		return nil, apierrors.ErrEmailAlreadyInUse
	}

	// hash with bcrypt
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apierrors.ErrInternalServer
	}

	user := &domain.User{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		PasswordHash: string(hashedBytes),
		SIN:          req.SIN,
		Status:       "PENDING",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Save(user); err != nil {
		return nil, apierrors.ErrInternalServer
	}

	return user, nil
}

func (s *identityService) Login(req LoginRequest) (*LoginResponse, *apierrors.APIError) {

	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, apierrors.ErrInternalServer
	}
	if user == nil {
		return nil, apierrors.ErrInvalidCredentials
	}

	// check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apierrors.ErrInvalidCredentials
	}

	// generate MFA Token (maximum 5 minutes)
	token, err := auth.GenerateToken(user.ID, user.Email, "PRE_AUTH", 5*time.Minute /* 5 minutes */)
	if err != nil {
		return nil, apierrors.ErrInternalServer
	}

	return &LoginResponse{
		Message:  "MFA code sent (simulated). Please verify.",
		MFAToken: token,
	}, nil
}

func (s *identityService) VerifyMFA(req VerifyMFARequest) (*TokenResponse, *apierrors.APIError) {
	claims, err := auth.ValidateToken(req.MFAToken)
	if err != nil || claims.Role != "PRE_AUTH" {
		return nil, apierrors.ErrInvalidMFA
	}

	// Verify simulated code
	if req.Code != "123456" && req.Code != "111111" {
		return nil, apierrors.ErrInvalidMFA
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil || user == nil {
		return nil, apierrors.ErrUserNotFound
	}

	// generate real access token (3 hours)
	token, err := auth.GenerateToken(user.ID, user.Email, "USER", 3*time.Hour)
	if err != nil {
		return nil, apierrors.ErrInternalServer
	}

	// save session in DB
	session := &domain.Session{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}
	if err := s.sessionRepo.Save(session); err != nil {
		return nil, apierrors.ErrInternalServer
	}

	return &TokenResponse{
		AccessToken: token,
	}, nil
}
