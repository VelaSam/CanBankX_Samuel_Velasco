package adapters

import (
	"errors"

	"canbankx/internal/identity/domain"
	"canbankx/internal/identity/ports"

	"gorm.io/gorm"
)

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) ports.SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Save(session *domain.Session) error {
	return r.db.Create(session).Error
}

func (r *sessionRepository) FindByToken(token string) (*domain.Session, error) {
	var session domain.Session
	err := r.db.Where("token = ?", token).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) DeleteByToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&domain.Session{}).Error
}
