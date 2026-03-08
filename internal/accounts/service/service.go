package service

import (
	"canbankx/internal/accounts/domain"
	"canbankx/internal/accounts/ports"
	apierrors "canbankx/internal/pkg/errors"
	"time"
)

type AccountService interface {
	OpenAccount(userID uint, req OpenAccountRequest) (*domain.Account, *apierrors.APIError)
	GetAccounts(userID uint) ([]domain.Account, *apierrors.APIError)
}

type OpenAccountRequest struct {
	Type string `json:"type" binding:"required,oneof=CHECKING SAVINGS"`
}

type accountService struct {
	repo ports.AccountRepository
}

func NewAccountService(repo ports.AccountRepository) AccountService {
	return &accountService{repo: repo}
}

func (s *accountService) OpenAccount(userID uint, req OpenAccountRequest) (*domain.Account, *apierrors.APIError) {
	account := &domain.Account{
		UserID:    userID,
		Type:      req.Type,
		Status:    "ACTIVE",
		Balance:   0.00,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Save(account); err != nil {
		return nil, apierrors.ErrInternalServer
	}

	return account, nil
}

func (s *accountService) GetAccounts(userID uint) ([]domain.Account, *apierrors.APIError) {
	accounts, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, apierrors.ErrInternalServer
	}
	return accounts, nil
}
