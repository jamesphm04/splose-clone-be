package repositories

import (
	"context"
	"errors"

	"github.com/jamesphm04/splose-clone-be/internal/models/entities"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type NoteRepository interface {
	Create(ctx context.Context, note *entities.Note) error
	FindByID(ctx context.Context, id string) (*entities.Note, error)
	FindByPatientID(ctx context.Context, patientID string) ([]entities.Note, error)
	List(ctx context.Context, offset, limit int) ([]entities.Note, int64, error)
	Update(ctx context.Context, note *entities.Note) error
	SoftDelete(ctx context.Context, id string) error
}

type noteRepo struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewNoteRepository returns a GORM-backed NoteRepository.
func NewNoteRepository(db *gorm.DB, log *zap.Logger) NoteRepository {
	return &noteRepo{
		db:  db,
		log: log.Named("note-repository"),
	}
}

func (r *noteRepo) Create(ctx context.Context, note *entities.Note) error {
	if err := r.db.WithContext(ctx).Create(note).Error; err != nil {
		r.log.Error("failed to create note", zap.String("noteID", note.ID), zap.Error(err))
	}

	r.log.Info("note created", zap.String("noteID", note.ID))
	return nil
}

func (r *noteRepo) FindByID(ctx context.Context, id string) (*entities.Note, error) {
	var n entities.Note
	err := r.db.
		WithContext(ctx).
		Preload("User").
		Preload("Patient").
		First(&n, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		r.log.Error("FindByID failed", zap.String("id", id), zap.Error(err))
	}

	return &n, nil
}

func (r *noteRepo) FindByPatientID(ctx context.Context, patientID string) ([]entities.Note, error) {
	var notes []entities.Note

	err := r.db.
		WithContext(ctx).
		Preload("User").
		Where("patient_id = ?", patientID).
		Find(&notes).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		r.log.Error("FindByPatientID failed", zap.String("patientID", patientID), zap.Error(err))
	}

	return notes, nil
}

func (r *noteRepo) List(ctx context.Context, offset, limit int) ([]entities.Note, int64, error) {
	var notes []entities.Note
	var total int64

	// count total
	if err := r.db.WithContext(ctx).Model(&entities.Note{}).Count(&total).Error; err != nil {
		r.log.Error("List count failed", zap.Error(err))
		return nil, 0, err
	}

	// list
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&notes).Error; err != nil {
		r.log.Error("List query failed", zap.Error(err))
		return nil, 0, err
	}

	return notes, total, nil
}

func (r *noteRepo) Update(ctx context.Context, note *entities.Note) error {
	if err := r.db.WithContext(ctx).Save(note).Error; err != nil {
		r.log.Error("Update failed", zap.String("noteID", note.ID), zap.Error(err))
		return err
	}

	return nil
}

func (r *noteRepo) SoftDelete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Delete(&entities.Note{}, "id = ?", id)
	if res.Error != nil {
		r.log.Error("SoftDelete failed", zap.String("noteID", id), zap.Error(res.Error))
	}

	if res.RowsAffected == 0 {
		return ErrNotFound
	}

	r.log.Info("note soft-deleted", zap.String("noteID", id))
	return nil
}
