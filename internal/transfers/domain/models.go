package domain

import (
	"time"
)

type Transfer struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	IdempotencyKey string    `gorm:"size:255;not null;uniqueIndex" json:"idempotencyKey"`
	SourceAccount  uint      `gorm:"not null;index" json:"sourceAccount"`
	TargetAccount  uint      `gorm:"not null;index" json:"targetAccount"`
	Amount         float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	Currency       string    `gorm:"size:3;not null;default:'CAD'" json:"currency"`
	Status         string    `gorm:"size:50;not null;default:'COMPLETED'" json:"status"` // PENDING, COMPLETED, FAILED
	CreatedAt      time.Time `json:"createdAt"`
}
