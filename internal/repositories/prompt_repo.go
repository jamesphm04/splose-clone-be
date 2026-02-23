package repositories

import (
	"context"
	"errors"

	"github.com/jamesphm04/splose-clone-be/internal/models/entities"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PromptRepository interface {
	Create(ctx context.Context, prompt *entities.Prompt) error
}

type promptRepo struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewPromptRepository returns a GORM-backed PromptRepository.
func NewPromptRepository(db *gorm.DB, log *zap.Logger) PromptRepository {
	return &promptRepo{
		db:  db,
		log: log.Named("prompt-repository"),
	}
}

func (r *promptRepo) Create(ctx context.Context, prompt *entities.Prompt) error {
	if err := r.db.WithContext(ctx).Create(prompt).Error; err != nil {
		r.log.Error("failed to create prompt", zap.String("promptID", prompt.ID), zap.Error(err))
		return err
	}

	r.log.Info("prompt created", zap.String("promptID", prompt.ID))
	return nil
}

func (r *promptRepo) FindByID(ctx context.Context, id string) (*entities.Prompt, error) {
	var p entities.Prompt
	err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		r.log.Error("FindByID failed", zap.String("id", id), zap.Error(err))
	}
	return &p, nil
}

func (r *promptRepo) List(ctx context.Context, offset, limit int) ([]entities.Prompt, int64, error) {
	var prompts []entities.Prompt
	var total int64

	// count total
	if err := r.db.WithContext(ctx).Model(&entities.Prompt{}).Count(&total).Error; err != nil {
		r.log.Error("List count failed", zap.Error(err))
		return nil, 0, err
	}

	// list
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&prompts).Error; err != nil {
		r.log.Error("List query failed", zap.Error(err))
		return nil, 0, err
	}

	return prompts, total, nil
}

func (r *promptRepo) Update(ctx context.Context, prompt *entities.Prompt) error {
	if err := r.db.WithContext(ctx).Save(prompt).Error; err != nil {
		r.log.Error("Update failed", zap.String("promptID", prompt.ID), zap.Error(err))
		return err
	}

	return nil
}

func (r *promptRepo) SoftDelete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Delete(&entities.Prompt{}, "id = ?", id)
	if res.Error != nil {
		r.log.Error("SoftDelete failed", zap.String("promptID", id), zap.Error(res.Error))
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}

	r.log.Info("prompt soft-deleted", zap.String("promptID", id))
	return nil
}
