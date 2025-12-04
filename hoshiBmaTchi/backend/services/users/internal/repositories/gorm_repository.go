package repositories

import (
	"context"

	"github.com/Hinsane5/hoshiBmaTchi/backend/services/users/internal/core/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type gormUserRepository struct{
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *gormUserRepository{
	err := db.AutoMigrate(&domain.User{}, &domain.Follow{})
    if err != nil {
    
    }
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

func (r *gormUserRepository) GetUserProfileWithStats(userID string) (*domain.User, int64, int64, error){
	var user domain.User
	if err := r.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, 0, 0, err
	}

	var followersCount int64
	r.db.Model(&domain.Follow{}).Where("following_id = ?", userID).Count(&followersCount)

	var followingCount int64
	r.db.Model(&domain.Follow{}).Where("follower_id = ?", userID).Count(&followingCount)

	return &user, followersCount, followingCount, nil
}

func (r *gormUserRepository) CreateFollow(followerID, followingID string) error {
    follow := &domain.Follow{
        FollowerID:  uuid.MustParse(followerID),
        FollowingID: uuid.MustParse(followingID),
    }
    return r.db.Create(follow).Error
}

func (r *gormUserRepository) DeleteFollow(followerID, followingID string) error {
    return r.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).
        Delete(&domain.Follow{}).Error
}

func (r *gormUserRepository) IsFollowing(followerID, followingID string) (bool, error) {
    var count int64
    err := r.db.Model(&domain.Follow{}).
        Where("follower_id = ? AND following_id = ?", followerID, followingID).
        Count(&count).Error
    return count > 0, err
}

func (r *gormUserRepository) GetFollowing(userID string) ([]string, error){
	var followingIDs []string
	var follows []domain.Follow

	err := r.db.Table("follows").
		Where("follower_id = ?", userID).
		Find(&follows).
		Error

	if err != nil {
		return nil, err
	}

	for _, follow := range follows {
		followingIDs = append(followingIDs, follow.FollowingID.String())
	}

	return followingIDs, nil
}

func (r *gormUserRepository) SearchUsers(ctx context.Context, query string) ([]*domain.User, error) {
    var users []*domain.User
    wildcard := "%" + query + "%"
    err := r.db.WithContext(ctx).
        Where("username ILIKE ? OR name ILIKE ?", wildcard, wildcard).
        Limit(20).
        Find(&users).Error
    
    if err != nil {
        return nil, err
    }
    return users, nil
}
