package service

import (
	"canbankx/internal/ledger/domain"
	"canbankx/internal/ledger/ports"
	apierrors "canbankx/internal/pkg/errors"
)

type LedgerService interface {
	GetAccountHistory(accountID uint) ([]domain.LedgerEntry, float64, *apierrors.APIError)
}

type ledgerService struct {
	repo ports.LedgerRepository
}

func NewLedgerService(repo ports.LedgerRepository) LedgerService {
	return &ledgerService{repo: repo}
}

func (s *ledgerService) GetAccountHistory(accountID uint) ([]domain.LedgerEntry, float64, *apierrors.APIError) {
	entries, err := s.repo.GetEntriesByAccountID(accountID)
	if err != nil {
		return nil, 0, apierrors.ErrInternalServer
	}

	balance, err := s.repo.CalculateBalance(accountID)
	if err != nil {
		return nil, 0, apierrors.ErrInternalServer
	}

	return entries, balance, nil
}
