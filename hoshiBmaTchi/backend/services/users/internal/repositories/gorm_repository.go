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

func (r *gormUserRepository) FindByEmail(email string) (*domain.User, error){
	var user domain.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil{
		return nil, err
	}

	return &user, nil
}

func (r *gormUserRepository) FindByEmailOrUsername(identifier string) (*domain.User, error){
	var user domain.User

	if err := r.db.Where("email = ? OR username = ?", identifier, identifier).First(&user).Error; err != nil{
		return nil, err
	}
	return &user, nil
}

func (r *gormUserRepository) FindByID(userID string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormUserRepository) UpdatePassword(userID string, newPassword string) error {
	result := r.db.Model(&domain.User{}).Where("id = ?", userID).Update("password", newPassword)
	return result.Error
}