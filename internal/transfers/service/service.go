package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"canbankx/internal/pkg/cache"

	ledgerDomain "canbankx/internal/ledger/domain"
	apierrors "canbankx/internal/pkg/errors"
	"canbankx/internal/transfers/domain"
	"canbankx/internal/transfers/ports"
)

type TransferService interface {
	ProcessTransfer(userID uint, req ProcessTransferRequest) (*domain.Transfer, *apierrors.APIError)
}

type ProcessTransferRequest struct {
	IdempotencyKey string  `json:"idempotencyKey" binding:"required"`
	TargetAccount  uint    `json:"targetAccount" binding:"required"`
	SourceAccount  uint    `json:"sourceAccount" binding:"required"`
	Amount         float64 `json:"amount" binding:"required,gt=0"`
}

type transferService struct {
	repo ports.TransferRepository
}

func NewTransferService(repo ports.TransferRepository) TransferService {
	return &transferService{repo: repo}
}

func (s *transferService) ProcessTransfer(userID uint, req ProcessTransferRequest) (*domain.Transfer, *apierrors.APIError) {
	//check idempotency
	existing, err := s.repo.FindByIdempotencyKey(req.IdempotencyKey)
	if err != nil {
		return nil, apierrors.ErrInternalServer
	}
	if existing != nil {
		return existing, nil
	}

	//prepare domain models
	transfer := &domain.Transfer{
		IdempotencyKey: req.IdempotencyKey,
		SourceAccount:  req.SourceAccount,
		TargetAccount:  req.TargetAccount,
		Amount:         req.Amount,
		Currency:       "CAD",
		Status:         "COMPLETED",
		CreatedAt:      time.Now(),
	}

	ledgerTx := &ledgerDomain.Transaction{
		Reference: fmt.Sprintf("TRF-%s", req.IdempotencyKey),
		Type:      "TRANSFER",
		CreatedAt: time.Now(),
	}

	entries := []ledgerDomain.LedgerEntry{
		{
			AccountID: req.SourceAccount,
			Amount:    -req.Amount,
			CreatedAt: time.Now(),
		},
		{
			AccountID: req.TargetAccount,
			Amount:    req.Amount,
			CreatedAt: time.Now(),
		},
	}

	//execute transactional transfer
	if err := s.repo.ExecuteTransfer(transfer, ledgerTx, entries); err != nil {
		if err.Error() == "insufficient funds" {
			return nil, &apierrors.APIError{ErrorCode: "BAD_REQUEST", Message: err.Error()}
		}
		if err.Error() == "source account not found" || err.Error() == "target account not found" {
			return nil, &apierrors.APIError{ErrorCode: "BAD_REQUEST", Message: err.Error()}
		}
		return nil, apierrors.ErrInternalServer
	}

	ctx := context.Background()
	sourceCacheKey := fmt.Sprintf("account:%d:history", req.SourceAccount)
	targetCacheKey := fmt.Sprintf("account:%d:history", req.TargetAccount)
	if err := cache.Client.Del(ctx, sourceCacheKey, targetCacheKey).Err(); err != nil {
		slog.Error("Failed to invalidate cache", "error", err)
	}

	return transfer, nil
}
