package repositories

import (
	"context"
	"errors"

	"github.com/jamesphm04/splose-clone-be/internal/models/entities"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ConversationRepository interface {
	Create(ctx context.Context, conversation *entities.Conversation) error
	FindByID(ctx context.Context, id string) (*entities.Conversation, error)
	List(ctx context.Context, offset, limit int) ([]entities.Conversation, int64, error)
	Update(ctx context.Context, conversation *entities.Conversation) error
	SoftDelete(ctx context.Context, id string) error
}

type conversationRepo struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewConversationRepository returns a GORM-backed ConversationRepository.
func NewConversationRepository(db *gorm.DB, log *zap.Logger) ConversationRepository {
	return &conversationRepo{
		db:  db,
		log: log.Named("conversation-repository"),
	}
}

func (r *conversationRepo) Create(ctx context.Context, conversation *entities.Conversation) error {
	if err := r.db.WithContext(ctx).Create(conversation).Error; err != nil {
		r.log.Error("failed to create conversation", zap.String("conversationID", conversation.ID), zap.Error(err))
	}
	r.log.Info("conversation created", zap.String("conversationID", conversation.ID))
	return nil
}

func (r *conversationRepo) FindByID(ctx context.Context, id string) (*entities.Conversation, error) {
	var c entities.Conversation
	err := r.db.WithContext(ctx).First(&c, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		r.log.Error("FindByID failed", zap.String("id", id), zap.Error(err))
	}
	return &c, nil
}

func (r *conversationRepo) List(ctx context.Context, offset, limit int) ([]entities.Conversation, int64, error) {
	var conversations []entities.Conversation
	var total int64

	// count total
	if err := r.db.WithContext(ctx).Model(&entities.Conversation{}).Count(&total).Error; err != nil {
		r.log.Error("List count failed", zap.Error(err))
		return nil, 0, err
	}

	// list
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&conversations).Error; err != nil {
		r.log.Error("List query failed", zap.Error(err))
		return nil, 0, err
	}

	return conversations, total, nil
}

func (r *conversationRepo) Update(ctx context.Context, conversation *entities.Conversation) error {
	if err := r.db.WithContext(ctx).Save(conversation).Error; err != nil {
		r.log.Error("Update failed", zap.String("conversationID", conversation.ID), zap.Error(err))
		return err
	}

	return nil
}

func (r *conversationRepo) SoftDelete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Delete(&entities.Conversation{}, "id = ?", id)
	if res.Error != nil {
		r.log.Error("SoftDelete failed", zap.String("conversationID", id), zap.Error(res.Error))
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}

	r.log.Info("conversation soft-deleted", zap.String("conversationID", id))
	return nil
}
