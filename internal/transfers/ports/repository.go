package ports

import (
	ledgerDomain "canbankx/internal/ledger/domain"
	"canbankx/internal/transfers/domain"
)

type TransferRepository interface {
	FindByIdempotencyKey(key string) (*domain.Transfer, error)

	ExecuteTransfer(
		transfer *domain.Transfer,
		ledgerTx *ledgerDomain.Transaction,
		entries []ledgerDomain.LedgerEntry,
	) error
}
