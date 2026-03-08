package ports

import "canbankx/internal/ledger/domain"

type LedgerRepository interface {
	SaveTransactionWithEntries(transaction *domain.Transaction, entries []domain.LedgerEntry) error
	GetEntriesByAccountID(accountID uint) ([]domain.LedgerEntry, error)
	CalculateBalance(accountID uint) (float64, error)
}
