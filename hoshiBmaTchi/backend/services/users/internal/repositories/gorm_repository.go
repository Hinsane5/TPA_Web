package repositories

import (
	"context"
	"time"

	"github.com/Hinsane5/hoshiBmaTchi/backend/services/users/internal/core/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type gormUserRepository struct{
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *gormUserRepository{
	err := db.AutoMigrate(
		&domain.User{}, 
		&domain.Follow{}, 
		&domain.Block{},
		&domain.CloseFriend{},
		&domain.HiddenStoryViewer{},
		&domain.VerificationRequest{},
	)
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

func (r *gormUserRepository) SearchUsers(ctx context.Context, query string, userID string) ([]*domain.User, error) {
    var users []*domain.User
    wildcard := "%" + query + "%"
    
    blockedSubQuery := r.db.Table("blocks").Select("blocked_id").Where("blocker_id = ?", userID)
    
    blockerSubQuery := r.db.Table("blocks").Select("blocker_id").Where("blocked_id = ?", userID)

    err := r.db.WithContext(ctx).
        Where("username ILIKE ? OR name ILIKE ?", wildcard, wildcard).
        Where("id != ?", userID).
        Where("id NOT IN (?)", blockedSubQuery).
        Where("id NOT IN (?)", blockerSubQuery).
        Limit(20).
        Find(&users).Error
    
    if err != nil {
        return nil, err
    }
    return users, nil
}

func (r *gormUserRepository) GetSuggestedUsers(ctx context.Context, userID string) ([]*domain.User, error) {
    var users []*domain.User
    
    // 1. Exclude people you already follow
    followingSubQuery := r.db.Table("follows").Select("following_id").Where("follower_id = ?", userID)
    
    // 2. Exclude people YOU blocked
    blockedSubQuery := r.db.Table("blocks").Select("blocked_id").Where("blocker_id = ?", userID)
    
    // 3. Exclude people who blocked YOU
    blockerSubQuery := r.db.Table("blocks").Select("blocker_id").Where("blocked_id = ?", userID)

    // Execute Main Query
    err := r.db.WithContext(ctx).
        Where("id != ?", userID).               
        Where("id NOT IN (?)", followingSubQuery).       
        Where("id NOT IN (?)", blockedSubQuery). 
        Where("id NOT IN (?)", blockerSubQuery).
        Order("RANDOM()").                      
        Limit(5).                               
        Find(&users).Error

    if err != nil {
        return nil, err
    }

    return users, nil
}


func (r *gormUserRepository) GetFollowingUsers(userID string) ([]*domain.User, error) {
	var users []*domain.User
	err := r.db.Joins("JOIN follows ON follows.following_id = users.id").
		Where("follows.follower_id = ?", userID).
		Find(&users).Error

	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *gormUserRepository) CreateBlock(blockerID, blockedID string) error {
    // 1. Create Block Record
    block := &domain.Block{
        BlockerID: uuid.MustParse(blockerID),
        BlockedID: uuid.MustParse(blockedID),
    }
    if err := r.db.Create(block).Error; err != nil {
        return err
    }

    r.db.Where("(follower_id = ? AND following_id = ?) OR (follower_id = ? AND following_id = ?)", 
        blockerID, blockedID, blockedID, blockerID).Delete(&domain.Follow{})
        
    return nil
}

func (r *gormUserRepository) DeleteBlock(blockerID, blockedID string) error {
    return r.db.Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).
        Delete(&domain.Block{}).Error
}

func (r *gormUserRepository) GetBlockedUsers(userID string) ([]*domain.User, error) {
    var users []*domain.User
    err := r.db.Joins("JOIN blocks ON blocks.blocked_id = users.id").
        Where("blocks.blocker_id = ?", userID).
        Find(&users).Error
    return users, err
}

func (r *gormUserRepository) IsBlocked(userA, userB string) (bool, error) {
    var count int64
    err := r.db.Model(&domain.Block{}).
        Where("(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)", 
        userA, userB, userB, userA).
        Count(&count).Error
    return count > 0, err
}

// --- Close Friends Operations ---
func (r *gormUserRepository) AddCloseFriend(userID, targetID uuid.UUID) error {
	cf := domain.CloseFriend{
		ID:            uuid.New(),
		UserID:        userID,
		CloseFriendID: targetID,
		CreatedAt:     time.Now(),
	}
	return r.db.Create(&cf).Error
}

func (r *gormUserRepository) RemoveCloseFriend(userID, targetID uuid.UUID) error {
	return r.db.Where("user_id = ? AND close_friend_id = ?", userID, targetID).Delete(&domain.CloseFriend{}).Error
}

func (r *gormUserRepository) GetCloseFriends(userID uuid.UUID) ([]domain.User, error) {
	var users []domain.User
	err := r.db.Joins("JOIN close_friends on close_friends.close_friend_id = users.id").
		Where("close_friends.user_id = ?", userID).
		Find(&users).Error
	return users, err
}

// --- Hidden Story Operations ---

func (r *gormUserRepository) HideStoryFromUser(userID, targetID uuid.UUID) error {
	hide := domain.HiddenStoryViewer{
		ID:           uuid.New(),
		UserID:       userID,
		HiddenUserID: targetID,
		CreatedAt:    time.Now(),
	}
	return r.db.Create(&hide).Error
}

func (r *gormUserRepository) UnhideStoryFromUser(userID, targetID uuid.UUID) error {
	return r.db.Where("user_id = ? AND hidden_user_id = ?", userID, targetID).Delete(&domain.HiddenStoryViewer{}).Error
}

func (r *gormUserRepository) GetHiddenStoryUsers(userID uuid.UUID) ([]domain.User, error) {
	var users []domain.User
	err := r.db.Joins("JOIN hidden_story_viewers on hidden_story_viewers.hidden_user_id = users.id").
		Where("hidden_story_viewers.user_id = ?", userID).
		Find(&users).Error
	return users, err
}

// --- Verification Operations ---

func (r *gormUserRepository) CreateVerificationRequest(req *domain.VerificationRequest) error {
	// ID generation logic in GORM BeforeCreate hooks is safer, but redundant explicit assignment is fine
	if req.ID == uuid.Nil {
		req.ID = uuid.New()
	}
	req.CreatedAt = time.Now()
	return r.db.Create(req).Error
}

func (r *gormUserRepository) UpdateUser(user *domain.User) error {
	return r.db.Save(user).Error
}

