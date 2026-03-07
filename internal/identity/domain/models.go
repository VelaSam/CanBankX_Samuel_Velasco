package domain

import (
	"time"

	"gorm.io/gorm"
)

// au cas ou tu oublie, les struct dans ce fichier sont les entités du domaine et sont utilisées par les repositories pour stocker les données dans la bd

type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	FirstName    string         `gorm:"size:100;not null" json:"firstName"`
	LastName     string         `gorm:"size:100;not null" json:"lastName"`
	Email        string         `gorm:"size:255;not null;uniqueIndex" json:"email"`
	PasswordHash string         `gorm:"not null" json:"-"`
	SIN          string         `gorm:"size:15;not null" json:"-"`                        // Social Insurance Number (Simulated)
	Status       string         `gorm:"size:50;not null;default:'PENDING'" json:"status"` // PENDING, ACTIVE, SUSPENDED
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type Session struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"userId"`
	Token     string    `gorm:"size:500;not null;index" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

type KYCDocument struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null" json:"userId"`
	DocumentID string    `gorm:"size:100;not null" json:"documentId"`
	Type       string    `gorm:"size:50;not null" json:"type"` // PASSPORT, DRIVERS_LICENSE
	VerifiedAt time.Time `json:"verifiedAt"`
	CreatedAt  time.Time `json:"createdAt"`
}
