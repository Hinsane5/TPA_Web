package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Hinsane5/hoshiBmaTchi/backend/services/posts/internal/core/domain"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/posts/internal/core/ports" // <--- ADD THIS IMPORT
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/posts/internal/core/services"
)

// --- Mock Implementation of PostRepository ---

type MockPostRepository struct {
	mock.Mock
}

// Methods actually used in the test
func (m *MockPostRepository) GetPostByID(ctx context.Context, postID string) (*domain.Post, error) {
	args := m.Called(ctx, postID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Post), args.Error(1)
}

func (m *MockPostRepository) DeletePost(ctx context.Context, postID string) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

// --- Stubs for other interface methods (Required to satisfy the interface) ---

// <--- ADD THIS MISSING METHOD --->
func (m *MockPostRepository) SearchHashtags(ctx context.Context, query string) ([]ports.HashtagSearchParam, error) {
	return nil, nil
}

// <--- ENSURE THIS IS PRESENT --->
func (m *MockPostRepository) CreateFullPost(ctx context.Context, post *domain.Post, mentions []domain.UserMention) error {
	return nil
}

func (m *MockPostRepository) CreatePost(ctx context.Context, post *domain.Post) error {
	return nil
}

func (m *MockPostRepository) GetPostsByUserID(ctx context.Context, userID string) ([]*domain.Post, error) {
	return nil, nil
}

func (m *MockPostRepository) LikePost(ctx context.Context, like *domain.PostLike) error {
	return nil
}

func (m *MockPostRepository) UnlikePost(ctx context.Context, userID, postID string) error {
	return nil
}

func (m *MockPostRepository) CreateComment(ctx context.Context, comment *domain.PostComment) error {
	return nil
}

func (m *MockPostRepository) GetCommentsForPost(ctx context.Context, postID string) ([]*domain.PostComment, error) {
	return nil, nil
}

func (m *MockPostRepository) GetFeedPosts(ctx context.Context, userIDs []string, currentUserID string, limit, offset int) ([]*domain.Post, error) {
	return nil, nil
}

func (m *MockPostRepository) CreateCollection(ctx context.Context, collection *domain.Collection) error {
	return nil
}

func (m *MockPostRepository) GetUserCollections(ctx context.Context, userID string) ([]*domain.Collection, error) {
	return nil, nil
}

func (m *MockPostRepository) ToggleSavePost(ctx context.Context, userID, postID, collectionID string) (bool, error) {
	return false, nil
}

func (m *MockPostRepository) CreatePostWithMentions(ctx context.Context, post *domain.Post, mentions []domain.UserMention) error {
	return nil
}

func (m *MockPostRepository) GetPostsByMention(ctx context.Context, targetUserID string, limit, offset int) ([]domain.Post, error) {
	return nil, nil
}

func (m *MockPostRepository) GetReels(ctx context.Context, limit, offset int) ([]*domain.Post, error) {
	return nil, nil
}

func (m *MockPostRepository) GetExplorePosts(ctx context.Context, limit, offset int, hashtag string) ([]*domain.Post, error) {
	return nil, nil
}

func (m *MockPostRepository) ToggleLike(ctx context.Context, postID string, userID string) (bool, error) {
	return false, nil
}

func (m *MockPostRepository) GetReelsByUserID(ctx context.Context, userID string) ([]*domain.Post, error) {
	return nil, nil
}

func (m *MockPostRepository) GetCollectionPosts(ctx context.Context, collectionID string, limit, offset int) ([]*domain.Post, error) {
	return nil, nil
}

func (m *MockPostRepository) UpdateCollection(ctx context.Context, collectionID, name, userID string) (*domain.Collection, error) {
	return nil, nil
}

func (m *MockPostRepository) DeleteCollection(ctx context.Context, collectionID, userID string) error {
	return nil
}

func (m *MockPostRepository) IsPostLikedByUser(ctx context.Context, postID, userID string) (bool, error) {
	return false, nil
}

func (m *MockPostRepository) GetPendingPostReports(ctx context.Context) ([]*domain.PostReport, error) {
	return nil, nil
}

func (m *MockPostRepository) GetPostReportByID(ctx context.Context, reportID string) (*domain.PostReport, error) {
	return nil, nil
}

func (m *MockPostRepository) UpdatePostReportStatus(ctx context.Context, reportID string, status string) error {
	return nil
}

func (m *MockPostRepository) CreatePostReport(report *domain.PostReport) error {
	return nil
}

// --- Unit Tests ---

func TestDeletePost(t *testing.T) {
	// Setup
	mockRepo := new(MockPostRepository)
	// We pass nil for amqpChan and userClient because DeletePost doesn't use them
	service := services.NewPostService(mockRepo, nil, nil)

	ctx := context.Background()
	postID := uuid.New()
	ownerID := uuid.New()
	otherUserID := uuid.New()

	// Create a dummy post object representing an existing post in DB
	existingPost := &domain.Post{
		ID:     postID,
		UserID: ownerID,
	}

	t.Run("Success: Owner deletes their own post", func(t *testing.T) {
		// Mock Expectation: GetPostByID called with postID, returns the post
		mockRepo.On("GetPostByID", ctx, postID.String()).Return(existingPost, nil).Once()

		// Mock Expectation: DeletePost called with postID, returns no error
		mockRepo.On("DeletePost", ctx, postID.String()).Return(nil).Once()

		// Execute
		err := service.DeletePost(ctx, postID.String(), ownerID.String())

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure: Non-owner tries to delete post", func(t *testing.T) {
		// Mock Expectation: GetPostByID called, returns the post
		mockRepo.On("GetPostByID", ctx, postID.String()).Return(existingPost, nil).Once()

		// Mock Expectation: DeletePost should NOT be called
		// We verify this implicitly by not adding an .On() for it, or checking AssertNotCalled later

		// Execute
		err := service.DeletePost(ctx, postID.String(), otherUserID.String())

		// Assert
		assert.Error(t, err)
		assert.Equal(t, "unauthorized: you are not the owner of this post", err.Error())
		
		// Ensure DeletePost was never called
		mockRepo.AssertNotCalled(t, "DeletePost")
	})

	t.Run("Failure: Post not found", func(t *testing.T) {
		// Mock Expectation: GetPostByID returns an error (e.g. record not found)
		mockRepo.On("GetPostByID", ctx, postID.String()).Return(nil, errors.New("record not found")).Once()

		// Execute
		err := service.DeletePost(ctx, postID.String(), ownerID.String())

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "post not found")
		
		// Ensure DeletePost was never called
		mockRepo.AssertNotCalled(t, "DeletePost")
	})
}