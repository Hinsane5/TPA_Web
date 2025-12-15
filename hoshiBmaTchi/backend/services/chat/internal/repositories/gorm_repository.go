package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/Hinsane5/hoshiBmaTchi/backend/services/chat/internal/core/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) SaveMessage(ctx context.Context, msg *domain.Message) error {
	return r.db.Create(msg).Error
}

func (r *ChatRepository) UnsendMessage(ctx context.Context, messageID, userID string) error {
	var msg domain.Message
	if err := r.db.First(&msg, "id = ?", messageID).Error; err != nil {
		return err
	}

	if msg.SenderID.String() != userID {
		return errors.New("unauthorized")
	}

	if time.Since(msg.CreatedAt) > 1*time.Minute {
		return errors.New("cannot unsend message older than 1 minute")
	}

	msg.IsUnsent = true
	msg.Content = "This message was unsent"
	msg.MediaURL = ""
	return r.db.Save(&msg).Error
}

func (r *ChatRepository) CreateConversation(ctx context.Context, conv *domain.Conversation, userIDs []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(conv).Error; err != nil {
			return err
		}
		
		for _, uid := range userIDs {
			parsedUUID, err := uuid.Parse(uid)
			if err != nil {
				continue 
			}

			p := domain.Participant{
				ConversationID: conv.ID,
				UserID:         parsedUUID,
				JoinedAt:       time.Now(),
			}
			if err := tx.Create(&p).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ChatRepository) GetConversations(ctx context.Context, userID string) ([]domain.Conversation, error) {
	var conversations []domain.Conversation
	
	subQuery := r.db.Table("participants").Select("conversation_id").Where("user_id = ?", userID)

	err := r.db.Preload("Participants").
		Where("id IN (?)", subQuery).
		Order("created_at DESC").
		Find(&conversations).Error

	return conversations, err
}

func (r *ChatRepository) GetMessageHistory(ctx context.Context, conversationID string, limit, offset int) ([]domain.Message, error) {
	var messages []domain.Message
	
	err := r.db.Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
		
	return messages, err
}

func (r *ChatRepository) AddParticipant(ctx context.Context, conversationID, userID string) error {

	cID, err := uuid.Parse(conversationID)
	if err != nil { return err }
	uID, err := uuid.Parse(userID)
	if err != nil { return err }

	p := domain.Participant{
		ConversationID: cID,
		UserID:         uID,
		JoinedAt:       time.Now(),
	}
	return r.db.Create(&p).Error
}

func (r *ChatRepository) RemoveParticipant(ctx context.Context, conversationID, userID string) error {
	return r.db.Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Delete(&domain.Participant{}).Error
}

func (r *ChatRepository) MarkConversationAsRead(ctx context.Context, conversationID, userID, messageID string) error {
	return r.db.Model(&domain.Participant{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Update("last_read_message_id", messageID).Error
}

func (r *ChatRepository) SearchMessages(ctx context.Context, conversationID, query string) ([]domain.Message, error) {
	var messages []domain.Message
	err := r.db.Where("conversation_id = ? AND content ILIKE ?", conversationID, "%"+query+"%").
		Order("created_at DESC").
		Find(&messages).Error
	return messages, err
}

func (r *ChatRepository) FindDirectConversation(ctx context.Context, user1ID, user2ID string) (*domain.Conversation, error) {
	var conversation domain.Conversation
	
	err := r.db.Table("conversations").
		Joins("JOIN participants p1 ON p1.conversation_id = conversations.id").
		Joins("JOIN participants p2 ON p2.conversation_id = conversations.id").
		Where("conversations.is_group = ?", false).
		Where("p1.user_id = ? AND p2.user_id = ?", user1ID, user2ID).
		First(&conversation).Error

	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (r *ChatRepository) DeleteConversation(ctx context.Context, conversationID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("conversation_id = ?", conversationID).Delete(&domain.Message{}).Error; err != nil {
			return err
		}

		if err := tx.Where("conversation_id = ?", conversationID).Delete(&domain.Participant{}).Error; err != nil {
			return err
		}

		if err := tx.Where("id = ?", conversationID).Delete(&domain.Conversation{}).Error; err != nil {
			return err
		}

		return nil
	})
}