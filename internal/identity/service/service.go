package service

import (
	"canbankx/internal/identity/domain"
	"canbankx/internal/identity/ports"
	apierrors "canbankx/internal/pkg/errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type IdentityService interface {
	RegisterUser(req RegisterRequest) (*domain.User, *apierrors.APIError)
	// TODO: additions like Login and VerifyKYC later
}

type RegisterRequest struct {
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	SIN       string `json:"sin" binding:"required,len=9"`
}

type identityService struct {
	userRepo ports.UserRepository
}

func NewIdentityService(repo ports.UserRepository) IdentityService {
	return &identityService{
		userRepo: repo,
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

	// save to repository
	if err := s.userRepo.Save(user); err != nil {
		return nil, apierrors.ErrInternalServer
	}

	return user, nil
}
