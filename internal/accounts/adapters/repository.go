package adapters

import (
	"errors"

	"canbankx/internal/accounts/domain"
	"canbankx/internal/accounts/ports"

	"gorm.io/gorm"
)

type accountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) ports.AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) Save(account *domain.Account) error {
	return r.db.Create(account).Error
}

func (r *accountRepository) FindByUserID(userID uint) ([]domain.Account, error) {
	var accounts []domain.Account
	err := r.db.Where("user_id = ?", userID).Find(&accounts).Error
	return accounts, err
}

func (r *accountRepository) FindByID(id uint) (*domain.Account, error) {
	var account domain.Account
	err := r.db.First(&account, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

func (r *accountRepository) Update(account *domain.Account) error {
	return r.db.Save(account).Error
}
