package adapters

import (
	"errors"

	accountsDomain "canbankx/internal/accounts/domain"
	ledgerDomain "canbankx/internal/ledger/domain"
	"canbankx/internal/transfers/domain"
	"canbankx/internal/transfers/ports"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type transferRepository struct {
	db *gorm.DB
}

func NewTransferRepository(db *gorm.DB) ports.TransferRepository {
	return &transferRepository{db: db}
}

func (r *transferRepository) FindByIdempotencyKey(key string) (*domain.Transfer, error) {
	var transfer domain.Transfer
	err := r.db.Where("idempotency_key = ?", key).First(&transfer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &transfer, nil
}

func (r *transferRepository) ExecuteTransfer(transfer *domain.Transfer, ledgerTx *ledgerDomain.Transaction, entries []ledgerDomain.LedgerEntry) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var source accountsDomain.Account
		//Lock source acc to prevent race conditions
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&source, transfer.SourceAccount).Error; err != nil {
			return errors.New("source account not found")
		}

		var target accountsDomain.Account
		//Lock target acc to prevent race conditions
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&target, transfer.TargetAccount).Error; err != nil {
			return errors.New("target account not found")
		}

		//calidate if enough funds in the source acc
		if source.Balance < transfer.Amount {
			return errors.New("insufficient funds")
		}

		//update acc balances
		source.Balance -= transfer.Amount
		target.Balance += transfer.Amount
		if err := tx.Save(&source).Error; err != nil {
			return err
		}
		if err := tx.Save(&target).Error; err != nil {
			return err
		}

		//Create Ledger
		if err := tx.Create(ledgerTx).Error; err != nil {
			return err
		}

		for i := range entries {
			entries[i].TransactionID = ledgerTx.ID
		}
		if err := tx.Create(&entries).Error; err != nil {
			return err
		}

		//The Transfer record is created last
		if err := tx.Create(transfer).Error; err != nil {
			return err
		}

		return nil
	})
}
