package domain

import (
	"time"
)

type Transaction struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Reference string    `gorm:"size:255;not null;uniqueIndex" json:"reference"` //  Transfer ID or external ref
	Type      string    `gorm:"size:50;not null" json:"type"`                   // TRANSFER, DEPOSIT, WITHDRAWAL
	CreatedAt time.Time `json:"createdAt"`
}

type LedgerEntry struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TransactionID uint      `gorm:"not null;index" json:"transactionId"`
	AccountID     uint      `gorm:"not null;index" json:"accountId"`
	Amount        float64   `gorm:"type:decimal(15,2);not null" json:"amount"` // Positive for credit, Negative for debit
	CreatedAt     time.Time `json:"createdAt"`
}
