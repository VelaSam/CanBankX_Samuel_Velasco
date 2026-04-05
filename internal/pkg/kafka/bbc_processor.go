package kafka

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"gorm.io/gorm"

	ledgerDomain "canbankx/internal/ledger/domain"
	"canbankx/internal/pkg/bbc"
)

type BBCEvent struct {
	EventType     string  `json:"event_type"`
	AccountID     uint    `json:"account_id"`
	Amount        float64 `json:"amount"`
	IsOrigin      bool    `json:"is_origin"`
	TransactionID uint    `json:"transaction_id"`
}

//ProcessBBCEvent on va route le generic bbc event vers le bon handler pour faire la job
func ProcessBBCEvent(db *gorm.DB, eventRaw []byte) {
	var event BBCEvent
	if err := json.Unmarshal(eventRaw, &event); err != nil {
		slog.Error("Failed to parse BBC event payload", "error", err)
		return
	}

	slog.Info("Processing BBC Event", "type", event.EventType, "account", event.AccountID, "amount", event.Amount)

	switch event.EventType {
	case "CREATE_TRANSACTION_ORIGIN":
		//on deduct le money du compte origin
		handleAdjustment(db, event, -event.Amount)

	case "CREATE_TRANSACTION_DESTINATION":
		//on ajoute des fonds direct dans le compte du receveur
		handleAdjustment(db, event, event.Amount)

	case "COMPENSATE_ORIGIN":
		//la bank receveuse a failed faque faut push l'argent BACK a l'origin
		handleAdjustment(db, event, event.Amount)

	default:
		slog.Warn("Unknown BBC event type ignored", "type", event.EventType)
	}
}

//handleAdjustment de hit le accounts-api pis le local ledger single sided
func handleAdjustment(db *gorm.DB, event BBCEvent, signedAmount float64) {
	//first on va trigger l'ajustement du balance live ds accounts-api
	accountsURL := os.Getenv("ACCOUNTS_SERVICE_URL")
	if accountsURL == "" {
		accountsURL = "http://accounts-api:8082"
	}

	payload := map[string]interface{}{
		"account_id": event.AccountID,
		"amount":     signedAmount,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", accountsURL+"/internal/v1/accounts/adjust", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)

	status := "CONFIRMED"
	if err != nil || resp.StatusCode != http.StatusOK {
		slog.Error("BBC Adjustment failed at accounts-api", "event", event.EventType, "error", err)
		status = "FAILED"
	} else {
		//on log l'action direct dans note transactions-db pour les records (l'audit)
		_ = db.Transaction(func(tx *gorm.DB) error {
			ledgerTx := ledgerDomain.Transaction{
				Reference: fmt.Sprintf("BBC-%s-%d", event.EventType, event.TransactionID),
				Type:      "EXTERNAL_TRANSFER",
				CreatedAt: time.Now(),
			}

			if err := tx.Create(&ledgerTx).Error; err != nil {
				return err
			}

			entry := ledgerDomain.LedgerEntry{
				TransactionID: ledgerTx.ID,
				AccountID:     event.AccountID,
				Amount:        signedAmount,
				CreatedAt:     time.Now(),
			}
			return tx.Create(&entry).Error
		})
	}

	//y faut tjrs envoyer l'outcome au central bank sauf si c un COMPENSATE_ORIGIN vu que cest le end de la chaine
	if event.EventType != "COMPENSATE_ORIGIN" {
		err = bbc.ConfirmTransaction(event.AccountID, event.Amount, event.IsOrigin, event.TransactionID, status)
		if err != nil {
			slog.Error("Failed to confirm transaction back to bbc", "error", err)
		}
	}
}
