package repositories

import (
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/users/internal/core/domain"
	"gorm.io/gorm"
)

type gormUserRepository struct{
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *gormUserRepository{
	db.AutoMigrate(&domain.User{})
	return &gormUserRepository{db: db}
}

func (r *gormUserRepository) Save(user *domain.User) error{
	result := r.db.Create(user)
	return result.Error
}