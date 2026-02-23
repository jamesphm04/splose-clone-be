package repositories

import (
	"context"
	"errors"

	"github.com/jamesphm04/splose-clone-be/internal/models/entities"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MessageRepository interface {
	Create(ctx context.Context, message *entities.Message) error
}

type messageRepo struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewMessageRepository returns a GORM-backed MessageRepository.
func NewMessageRepository(db *gorm.DB, log *zap.Logger) MessageRepository {
	return &messageRepo{
		db:  db,
		log: log.Named("message-repository"),
	}
}

func (r *messageRepo) Create(ctx context.Context, message *entities.Message) error {
	if err := r.db.WithContext(ctx).Create(message).Error; err != nil {
		r.log.Error("failed to create message", zap.String("messageID", message.ID), zap.Error(err))
		return err
	}

	r.log.Info("message created", zap.String("messageID", message.ID))
	return nil
}

func (r *messageRepo) FindByID(ctx context.Context, id string) (*entities.Message, error) {
	var m entities.Message
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		r.log.Error("FindByID failed", zap.String("id", id), zap.Error(err))
	}
	return &m, nil
}

func (r *messageRepo) List(ctx context.Context, offset, limit int) ([]entities.Message, int64, error) {
	var messages []entities.Message
	var total int64

	// count total
	if err := r.db.WithContext(ctx).Model(&entities.Message{}).Count(&total).Error; err != nil {
		r.log.Error("List count failed", zap.Error(err))
		return nil, 0, err
	}

	// list
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&messages).Error; err != nil {
		r.log.Error("List query failed", zap.Error(err))
		return nil, 0, err
	}

	return messages, total, nil
}

func (r *messageRepo) Update(ctx context.Context, message *entities.Message) error {
	if err := r.db.WithContext(ctx).Save(message).Error; err != nil {
		r.log.Error("Update failed", zap.String("messageID", message.ID), zap.Error(err))
		return err
	}

	return nil
}

func (r *messageRepo) SoftDelete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Delete(&entities.Message{}, "id = ?", id)
	if res.Error != nil {
		r.log.Error("SoftDelete failed", zap.String("messageID", id), zap.Error(res.Error))
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}

	r.log.Info("message soft-deleted", zap.String("messageID", id))
	return nil
}
