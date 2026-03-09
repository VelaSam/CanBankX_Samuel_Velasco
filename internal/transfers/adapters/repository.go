package adapters

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	ledgerDomain "canbankx/internal/ledger/domain"
	"canbankx/internal/transfers/domain"
	"canbankx/internal/transfers/ports"

	"gorm.io/gorm"
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
	accountsURL := os.Getenv("ACCOUNTS_SERVICE_URL")
	if accountsURL == "" {
		accountsURL = "http://accounts-api:8082"
	}

	payload := map[string]interface{}{
		"sourceAccount": transfer.SourceAccount,
		"targetAccount": transfer.TargetAccount,
		"amount":        transfer.Amount,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", accountsURL+"/internal/v1/accounts/transfer", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errMsg, ok := errResp["error"].(string); ok {
			return errors.New(errMsg)
		}
		return errors.New("failed to transfer amounts in accounts API")
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ledgerTx).Error; err != nil {
			return err
		}

		for i := range entries {
			entries[i].TransactionID = ledgerTx.ID
		}
		if err := tx.Create(&entries).Error; err != nil {
			return err
		}

		if err := tx.Create(transfer).Error; err != nil {
			return err
		}

		return nil
	})
}
