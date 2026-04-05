package bbc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

var defaultBBCUrl = "http://bbc-service:5005"
var BankID = 2 // SAM-BANK is 2 based on 01_init_banks.sql

type InitiateTxReq struct {
	FromAccountID uint    `json:"from_account_id"`
	FromBankID    int     `json:"from_bank_id"`
	ToKey         string  `json:"to_key"`
	Amount        float64 `json:"amount"`
}

type ConfirmTxReq struct {
	FromAccountID uint    `json:"from_account_id"`
	FromBankID    int     `json:"from_bank_id"`
	Amount        float64 `json:"amount"`
	IsOrigin      bool    `json:"is_origin"`
	TransactionID uint    `json:"transaction_id"`
	Status        string  `json:"status"` //"CONFIRMED" ou "FAILED"
}

//initiate la transaction outward vers le central bank (bbc)
func InitiateTransaction(fromAcc int, toKey string, amount float64) error {
	payload := InitiateTxReq{
		FromAccountID: uint(fromAcc),
		FromBankID:    BankID,
		ToKey:         toKey,
		Amount:        amount,
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/transactions", defaultBBCUrl)
	slog.Info("Sending Request to BBC", "url", url, "payload", string(body))
	
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Failed to call BBC /transactions", "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("BBC rejected transaction with status %d", resp.StatusCode)
	}

	return nil
}

//on send ConfirmTransaction apres avoir process un event kafka dans l'engin local
func ConfirmTransaction(fromAcc uint, amount float64, isOrigin bool, transactionID uint, status string) error {
	payload := ConfirmTxReq{
		FromAccountID: fromAcc,
		FromBankID:    BankID,
		Amount:        amount,
		IsOrigin:      isOrigin,
		TransactionID: transactionID,
		Status:        status,
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/confirm-transaction", defaultBBCUrl)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Failed to confirm transaction with BBC", "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("BBC confirmation rejected with status %d", resp.StatusCode)
	}

	return nil
}
