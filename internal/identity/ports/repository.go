package ports

import "canbankx/internal/identity/domain"

type UserRepository interface {
	Save(user *domain.User) error
	FindByEmail(email string) (*domain.User, error)
	FindByID(id uint) (*domain.User, error)
	Update(user *domain.User) error
}

type SessionRepository interface {
	Save(session *domain.Session) error
	FindByToken(token string) (*domain.Session, error)
	DeleteByToken(token string) error
}
