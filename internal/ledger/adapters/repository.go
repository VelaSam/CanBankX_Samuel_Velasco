package adapters

import (
	"database/sql"
	"errors"

	"canbankx/internal/ledger/domain"
	"canbankx/internal/ledger/ports"

	"gorm.io/gorm"
)

type ledgerRepository struct {
	db *gorm.DB
}

func NewLedgerRepository(db *gorm.DB) ports.LedgerRepository {
	return &ledgerRepository{db: db}
}

func (r *ledgerRepository) SaveTransactionWithEntries(transaction *domain.Transaction, entries []domain.LedgerEntry) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(transaction).Error; err != nil {
			return err
		}
		for i := range entries {
			entries[i].TransactionID = transaction.ID
		}
		if err := tx.Create(&entries).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *ledgerRepository) GetEntriesByAccountID(accountID uint) ([]domain.LedgerEntry, error) {
	var entries []domain.LedgerEntry
	err := r.db.Where("account_id = ?", accountID).Order("created_at desc").Find(&entries).Error
	return entries, err
}

func (r *ledgerRepository) CalculateBalance(accountID uint) (float64, error) {
	var total sql.NullFloat64
	err := r.db.Model(&domain.LedgerEntry{}).
		Select("sum(amount)").
		Where("account_id = ?", accountID).
		Scan(&total).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0.0, nil
		}
		return 0.0, err
	}
	return total.Float64, nil
}
