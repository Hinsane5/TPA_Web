package repositories

import (
	"context"
	"log"
	"time"

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
    var post domain.Post
    // 1. Fetch Post with Media
    if err := r.db.Preload("Media").Where("id = ?", postID).First(&post).Error; err != nil {
        return nil, err
    }

    // 2. Count Likes
    var likesCount int64
    r.db.Model(&domain.PostLike{}).Where("post_id = ?", postID).Count(&likesCount)
    post.LikesCount = int32(likesCount)

    // 3. Count Comments
    var commentsCount int64
    r.db.Model(&domain.PostComment{}).Where("post_id = ?", postID).Count(&commentsCount)
    post.CommentsCount = int32(commentsCount)

    return &post, nil
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
    log.Printf("[DEBUG] GetPostsByMention called for TargetUserID: %s", targetUserID)    
    var posts []domain.Post
    
    err := r.db.Table("posts").
		Joins("JOIN user_mentions ON user_mentions.post_id = posts.id").
		Where("user_mentions.mentioned_user_id = ?", targetUserID).
		Order("posts.created_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("Media", func(db *gorm.DB) *gorm.DB {
            return db.Order("sequence asc")
        }).
		Find(&posts).Error

    if err != nil {
        log.Printf("[ERROR] Database query failed: %v", err)
        return nil, err
    }

    for i := range posts {
        var likes int64
        r.db.Model(&domain.PostLike{}).Where("post_id = ?", posts[i].ID).Count(&likes)
        posts[i].LikesCount = int32(likes)

        var comments int64
        r.db.Model(&domain.PostComment{}).Where("post_id = ?", posts[i].ID).Count(&comments)
        posts[i].CommentsCount = int32(comments)
        
    }

    log.Printf("[DEBUG] Query successful. Found %d posts for user %s", len(posts), targetUserID)
    return posts, nil
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

func (r *GormPostRepository) GetExplorePosts(ctx context.Context, limit, offset int, hashtag string) ([]*domain.Post, error) {
    var posts []*domain.Post
    
    query := r.db.WithContext(ctx).
        Preload("Media", func(db *gorm.DB) *gorm.DB {
            return db.Order("sequence asc")
        }).
        Order("created_at desc").
        Limit(limit).
        Offset(offset)

    if hashtag != "" {
        query = query.Where("caption ILIKE ?", "%#"+hashtag+"%") 
    }

    if err := query.Find(&posts).Error; err != nil {
        return nil, err
    }

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

func (r *GormPostRepository) ToggleLike(ctx context.Context, postID string, userID string) (bool, error) {
	var like domain.PostLike
	result := r.db.WithContext(ctx).Where("post_id = ? AND user_id = ?", postID, userID).First(&like)

	if result.Error == nil {
		if err := r.db.WithContext(ctx).Delete(&like).Error; err != nil {
			return false, err
		}
		return false, nil 
	} else if result.Error == gorm.ErrRecordNotFound {
		pID, _ := uuid.Parse(postID)
		uID, _ := uuid.Parse(userID)
		
		newLike := domain.PostLike{
			PostID: pID,
			UserID: uID,
		}
		
		if err := r.db.WithContext(ctx).Create(&newLike).Error; err != nil {
			return false, err
		}
		return true, nil 
	} else {
		return false, result.Error
	}
}

func (r *GormPostRepository) GetReelsByUserID(ctx context.Context, userID string) ([]*domain.Post, error) {
    var posts []*domain.Post

    err := r.db.WithContext(ctx).
        Preload("Media", func(db *gorm.DB) *gorm.DB {
            return db.Order("sequence asc")
        }).
        Where("user_id = ? AND is_reel = ?", userID, true).
        Order("created_at desc").
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

    }

    return posts, nil
}

func (r *GormPostRepository) GetCollectionPosts(ctx context.Context, collectionID string, limit, offset int) ([]*domain.Post, error) {
    var savedPosts []domain.SavedPost
    
    err := r.db.WithContext(ctx).
        Preload("Post.Media", func(db *gorm.DB) *gorm.DB {
            return db.Order("sequence asc")
        }).
        Where("collection_id = ?", collectionID).
        Order("created_at desc").
        Limit(limit).
        Offset(offset).
        Find(&savedPosts).Error

    if err != nil {
        return nil, err
    }

    var posts []*domain.Post
    for _, sp := range savedPosts {
        post := sp.Post
        posts = append(posts, &post)
    }
    
    for _, post := range posts {
         var likes int64
        r.db.Model(&domain.PostLike{}).Where("post_id = ?", post.ID).Count(&likes)
        post.LikesCount = int32(likes)

        var comments int64
        r.db.Model(&domain.PostComment{}).Where("post_id = ?", post.ID).Count(&comments)
        post.CommentsCount = int32(comments)
        
        post.IsLiked = false 
    }

    return posts, nil
}

func (r *GormPostRepository) UpdateCollection(ctx context.Context, collectionID, name, userID string) (*domain.Collection, error) {
    var collection domain.Collection
    
    if err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", collectionID, userID).First(&collection).Error; err != nil {
        return nil, err
    }

    collection.Name = name
    if err := r.db.WithContext(ctx).Save(&collection).Error; err != nil {
        return nil, err
    }

    return &collection, nil
}

func (r *GormPostRepository) DeleteCollection(ctx context.Context, collectionID, userID string) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        if err := tx.Where("collection_id = ?", collectionID).Delete(&domain.SavedPost{}).Error; err != nil {
            return err
        }

        result := tx.Where("id = ? AND user_id = ?", collectionID, userID).Delete(&domain.Collection{})
        if result.Error != nil {
            return result.Error
        }
        if result.RowsAffected == 0 {
            return gorm.ErrRecordNotFound
        }
        return nil
    })
}

func (r *GormPostRepository) IsPostLikedByUser(ctx context.Context, postID, userID string) (bool, error) {
    var count int64
    err := r.db.Model(&domain.PostLike{}).
        Where("post_id = ? AND user_id = ?", postID, userID).
        Count(&count).Error
    return count > 0, err
}

func (r *GormPostRepository) GetPendingPostReports(ctx context.Context) ([]*domain.PostReport, error) {
    var reports []*domain.PostReport
    err := r.db.WithContext(ctx).Where("status = ?", "PENDING").Find(&reports).Error
    return reports, err
}

func (r *GormPostRepository) UpdatePostReportStatus(ctx context.Context, reportID string, status string) error {
    return r.db.WithContext(ctx).Model(&domain.PostReport{}).Where("id = ?", reportID).Update("status", status).Error
}

func (r *GormPostRepository) DeletePost(ctx context.Context, postID string) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        // 1. Delete associated SavedPosts
        if err := tx.Where("post_id = ?", postID).Delete(&domain.SavedPost{}).Error; err != nil {
            return err
        }
        
        // 2. Delete associated Likes
        if err := tx.Where("post_id = ?", postID).Delete(&domain.PostLike{}).Error; err != nil {
            return err
        }

        // 3. Delete associated Comments
        if err := tx.Where("post_id = ?", postID).Delete(&domain.PostComment{}).Error; err != nil {
            return err
        }

        // 4. Delete associated Mentions (if any)
        if err := tx.Where("post_id = ?", postID).Delete(&domain.UserMention{}).Error; err != nil {
            return err
        }

        // 5. Delete associated Reports (since the post is gone, reports should be cleaned or resolved)
        if err := tx.Where("post_id = ?", postID).Delete(&domain.PostReport{}).Error; err != nil {
            return err
        }
        
        // 6. Delete the Post itself
        if err := tx.Where("id = ?", postID).Delete(&domain.Post{}).Error; err != nil {
            return err
        }

        return nil
    })
}

func (r *GormPostRepository) GetPostReportByID(ctx context.Context, reportID string) (*domain.PostReport, error) {
    var report domain.PostReport
    if err := r.db.WithContext(ctx).Where("id = ?", reportID).First(&report).Error; err != nil {
        return nil, err
    }
    return &report, nil
}

// Add this method
func (r *GormPostRepository) CreatePostReport(report *domain.PostReport) error {
    if report.ID == uuid.Nil {
        report.ID = uuid.New()
    }
    report.CreatedAt = time.Now()
    report.Status = "PENDING"
    return r.db.Create(report).Error
}