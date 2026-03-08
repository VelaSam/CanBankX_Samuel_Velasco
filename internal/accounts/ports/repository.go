package ports

import "canbankx/internal/accounts/domain"

type AccountRepository interface {
	Save(account *domain.Account) error
	FindByUserID(userID uint) ([]domain.Account, error)
	FindByID(id uint) (*domain.Account, error)
	Update(account *domain.Account) error
}
