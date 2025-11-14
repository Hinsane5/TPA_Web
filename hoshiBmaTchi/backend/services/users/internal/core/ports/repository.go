package ports

import "github.com/Hinsane5/hoshiBmaTchi/backend/services/users/internal/core/domain"

type UserRepository interface{
	Save(user *domain.User) error
	FindByEmail(email string) (*domain.User, error)
	FindByEmailOrUsername(identifier string) (*domain.User, error)
	FindByID(userID string) (*domain.User, error)
	UpdatePassword(userID string, newPassword string) error
}