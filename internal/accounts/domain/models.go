package domain

import (
	"time"

	"gorm.io/gorm"
)

type Account struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"userId"`
	Type      string         `gorm:"size:50;not null" json:"type"`                    // CHECKING, SAVINGS
	Status    string         `gorm:"size:50;not null;default:'ACTIVE'" json:"status"` // ACTIVE, CLOSED, FROZEN
	Balance   float64        `gorm:"type:decimal(15,2);default:0.00" json:"balance"`  // Snapshot of the current balance
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
