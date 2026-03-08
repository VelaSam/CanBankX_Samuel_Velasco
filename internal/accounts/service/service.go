package service

import (
	"canbankx/internal/accounts/domain"
	"canbankx/internal/accounts/ports"
	ledgerDomain "canbankx/internal/ledger/domain"
	ledgerService "canbankx/internal/ledger/service"
	apierrors "canbankx/internal/pkg/errors"
	"time"
)

type AccountService interface {
	OpenAccount(userID uint, req OpenAccountRequest) (*domain.Account, *apierrors.APIError)
	GetAccounts(userID uint) ([]domain.Account, *apierrors.APIError)
	GetAccountHistory(userID uint, accountID uint) ([]ledgerDomain.LedgerEntry, float64, *apierrors.APIError)
}

type OpenAccountRequest struct {
	Type string `json:"type" binding:"required,oneof=CHECKING SAVINGS"`
}

type accountService struct {
	repo          ports.AccountRepository
	ledgerService ledgerService.LedgerService
}

func NewAccountService(repo ports.AccountRepository, ledgerSvc ledgerService.LedgerService) AccountService {
	return &accountService{
		repo:          repo,
		ledgerService: ledgerSvc,
	}
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

func (s *accountService) GetAccountHistory(userID uint, accountID uint) ([]ledgerDomain.LedgerEntry, float64, *apierrors.APIError) {
	//before everything, verify that the account belongs to the user
	account, err := s.repo.FindByID(accountID)
	if err != nil || account == nil {
		return nil, 0, apierrors.ErrInternalServer
	}
	if account.UserID != userID {
		return nil, 0, &apierrors.APIError{ErrorCode: "FORBIDDEN", Message: "Not your account"}
	}

	return s.ledgerService.GetAccountHistory(accountID)
}
