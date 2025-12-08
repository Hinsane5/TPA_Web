package repositories

import (
	"context"

	"github.com/Hinsane5/hoshiBmaTchi/backend/services/posts/internal/core/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormPostRepository struct {
	db *gorm.DB
}

func NewGormPostRepository(db *gorm.DB) *GormPostRepository{
	return &GormPostRepository{db: db}
}

func (r *GormPostRepository) CreatePost(ctx context.Context, post *domain.Post) error {
	result := r.db.WithContext(ctx).Create(post)
    return result.Error 
}

func (r *GormPostRepository) GetPostByID(ctx context.Context, postID string) (*domain.Post, error) {
	return nil, nil 
}

func (r *GormPostRepository) GetPostsByUserID(ctx context.Context, userID string) ([]*domain.Post, error) {
	var posts []*domain.Post

	err := r.db.WithContext(ctx).
        Preload("Media", func(db *gorm.DB) *gorm.DB {
            return db.Order("sequence asc")
        }).
        Where("user_id = ?", userID).
        Order("created_at desc").
        Find(&posts).Error

	if err != nil {
		return nil, err
	}

	return posts, nil 
}

func (r *GormPostRepository) LikePost(ctx context.Context, like *domain.PostLike) error {
    result := r.db.WithContext(ctx).Create(like)
    return result.Error
}

func (r *GormPostRepository) UnlikePost(ctx context.Context, userID, postID string) error {
    result := r.db.WithContext(ctx).
        Where("user_id = ? AND post_id = ?", userID, postID).
        Delete(&domain.PostLike{})

    return result.Error
}

func (r *GormPostRepository) CreateComment(ctx context.Context, comment *domain.PostComment) error {
	result := r.db.WithContext(ctx).Create(comment)
	return result.Error
}

func (r *GormPostRepository) GetCommentsForPost(ctx context.Context, postID string) ([]*domain.PostComment, error){
	var comments []*domain.PostComment
	err := r.db.WithContext(ctx).Where("post_id = ?", postID).Order("created_at asc").Find(&comments).Error
	return comments, err
}

func (r *GormPostRepository) GetFeedPosts(ctx context.Context, userIDs []string, currentUserID string, limit, offset int) ([]*domain.Post, error) {
	var posts []*domain.Post
	
	err := r.db.WithContext(ctx).
        Preload("Media", func(db *gorm.DB) *gorm.DB {
            return db.Order("sequence asc")
        }).
        Where("user_id IN ?", userIDs).
        Order("created_at desc").
        Limit(limit).
        Offset(offset).
        Find(&posts).Error

	if err != nil {
		return nil, err
	}

	for _, post := range posts {
        var likes int64
        r.db.Model(&domain.PostLike{}).Where("post_id = ?", post.ID).Count(&likes)
        post.LikesCount = int32(likes)

        var comments int64
        r.db.Model(&domain.PostComment{}).Where("post_id = ?", post.ID).Count(&comments)
        post.CommentsCount = int32(comments)

		var isLikedCount int64
        r.db.Model(&domain.PostLike{}).
             Where("post_id = ? AND user_id = ?", post.ID, currentUserID).
             Count(&isLikedCount)
        
        post.IsLiked = isLikedCount > 0
    }

	return posts, nil
}

func (r *GormPostRepository) CreateCollection(ctx context.Context, collection *domain.Collection) error {
	return r.db.WithContext(ctx).Create(collection).Error
}

func (r *GormPostRepository) GetUserCollections(ctx context.Context, userID string) ([]*domain.Collection, error) {
	var collections []*domain.Collection
	
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&collections).Error; err != nil {
		return nil, err
	}

	for _, coll := range collections {
		var savedPosts []domain.SavedPost
		r.db.WithContext(ctx).
			Preload("Post.Media", func(db *gorm.DB) *gorm.DB {
                return db.Order("sequence asc")
            }).
			Where("collection_id = ?", coll.ID).
			Order("created_at desc").
			Limit(4).
			Find(&savedPosts)
		
		coll.SavedPosts = savedPosts
	}

	return collections, nil
}

func (r *GormPostRepository) ToggleSavePost(ctx context.Context, userID, postID, collectionID string) (bool, error) {
	var savedPost domain.SavedPost
	
	query := r.db.WithContext(ctx).Where("user_id = ? AND post_id = ?", userID, postID)
	
	if collectionID != "" {
		query = query.Where("collection_id = ?", collectionID)
	}

	result := query.First(&savedPost)

	if result.Error == nil {
		if err := r.db.WithContext(ctx).Delete(&savedPost).Error; err != nil {
			return true, err
		}
		return false, nil 
	}

	newSave := domain.SavedPost{
		UserID: uuid.MustParse(userID),
		PostID: uuid.MustParse(postID),
	}

	if collectionID != "" {
		newSave.CollectionID = uuid.MustParse(collectionID)
	} else {
		var defaultCollection domain.Collection
		
		err := r.db.WithContext(ctx).
			Where("user_id = ? AND name = ?", userID, "All Posts").
			First(&defaultCollection).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				defaultCollection = domain.Collection{
					UserID: uuid.MustParse(userID),
					Name:   "All Posts",
				}
				if createErr := r.db.WithContext(ctx).Create(&defaultCollection).Error; createErr != nil {
					return false, createErr
				}
			} else {
				return false, err 
			}
		}
		newSave.CollectionID = defaultCollection.ID
	}

	if err := r.db.WithContext(ctx).Create(&newSave).Error; err != nil {
		return false, err
	}

	return true, nil 
}

func (r *GormPostRepository) CreatePostWithMentions(ctx context.Context, post *domain.Post, mentions []domain.UserMention) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(post).Error; err != nil {
            return err
        }

        if len(mentions) > 0 {
            for i := range mentions {
                mentions[i].PostID = post.ID
            }
            if err := tx.Create(&mentions).Error; err != nil {
                return err
            }
        }

        return nil
    })
}

func (r *GormPostRepository) GetPostsByMention(ctx context.Context, targetUserID string, limit, offset int) ([]domain.Post, error) {
    var posts []domain.Post
    
    err := r.db.Table("posts").
		Joins("JOIN user_mentions ON user_mentions.post_id = posts.id").
		Where("user_mentions.mentioned_user_id = ?", targetUserID).
		Order("posts.created_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("Media").
		Find(&posts).Error

    return posts, err
}

// Add this function to the file
func (r *GormPostRepository) GetReels(ctx context.Context, limit, offset int) ([]*domain.Post, error) {
    var posts []*domain.Post
    
    err := r.db.WithContext(ctx).
        Preload("Media", func(db *gorm.DB) *gorm.DB {
            return db.Order("sequence asc")
        }).
        Where("is_reel = ?", true).
        Order("created_at desc").
        Limit(limit).
        Offset(offset).
        Find(&posts).Error

    if err != nil {
        return nil, err
    }

    // Populate likes/comments counts (same as Feed)
    for _, post := range posts {
        var likes int64
        r.db.Model(&domain.PostLike{}).Where("post_id = ?", post.ID).Count(&likes)
        post.LikesCount = int32(likes)

        var comments int64
        r.db.Model(&domain.PostComment{}).Where("post_id = ?", post.ID).Count(&comments)
        post.CommentsCount = int32(comments)
        
    }

    return posts, nil
}