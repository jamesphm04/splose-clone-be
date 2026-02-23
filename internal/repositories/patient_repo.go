package repositories

import (
	"context"
	"errors"

	"github.com/jamesphm04/splose-clone-be/internal/models/entities"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PatientRepository interface {
	Create(ctx context.Context, patient *entities.Patient) error
	FindByID(ctx context.Context, id string) (*entities.Patient, error)
	FindByEmail(ctx context.Context, email string) (*entities.Patient, error)
	FindByPhoneNumber(ctx context.Context, phoneNumber string) (*entities.Patient, error)
	List(ctx context.Context, offset, limit int) ([]entities.Patient, int64, error)
	Update(ctx context.Context, patient *entities.Patient) error
	SoftDelete(ctx context.Context, id string) error
}

type patientRepo struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewPatientRepository returns a GORM-backed PatientRepository
func NewPatientRepository(db *gorm.DB, log *zap.Logger) PatientRepository {
	return &patientRepo{
		db:  db,
		log: log.Named("patient-repository"),
	}
}

func (r *patientRepo) Create(ctx context.Context, patient *entities.Patient) error {
	if err := r.db.WithContext(ctx).Create(patient).Error; err != nil {
		r.log.Error("failed to create patient", zap.String("email", patient.Email), zap.Error(err))
		return err
	}

	r.log.Info("patient created", zap.String("email", patient.Email))
	return nil
}

func (r *patientRepo) FindByID(ctx context.Context, id string) (*entities.Patient, error) {
	var p entities.Patient
	err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		r.log.Error("FindByID failed", zap.String("id", id), zap.Error(err))
	}

	return &p, nil
}

func (r *patientRepo) FindByEmail(ctx context.Context, email string) (*entities.Patient, error) {
	var p entities.Patient
	err := r.db.WithContext(ctx).First(&p, "email = ?", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		r.log.Error("FindByEmail failed", zap.String("email", email), zap.Error(err))
	}
	return &p, nil
}

func (r *patientRepo) FindByPhoneNumber(ctx context.Context, phoneNumber string) (*entities.Patient, error) {
	var p entities.Patient
	err := r.db.WithContext(ctx).First(&p, "phoneNumber = ?", phoneNumber).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		r.log.Error("FindByPhoneNumber failed", zap.String("phoneNumber", phoneNumber), zap.Error(err))
	}
	return &p, nil
}

func (r *patientRepo) List(ctx context.Context, offset, limit int) ([]entities.Patient, int64, error) {
	var patients []entities.Patient
	var total int64

	// count total
	if err := r.db.WithContext(ctx).Model(&entities.Patient{}).Count(&total).Error; err != nil {
		r.log.Error("List count failed", zap.Error(err))
		return nil, 0, err
	}

	// list
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&patients).Error; err != nil {
		r.log.Error("List query failed", zap.Error(err))
		return nil, 0, err
	}

	return patients, total, nil
}

func (r *patientRepo) Update(ctx context.Context, patient *entities.Patient) error {
	if err := r.db.WithContext(ctx).Save(patient).Error; err != nil {
		r.log.Error("Update failed", zap.String("patientID", patient.ID), zap.Error(err))
		return err
	}

	return nil
}

func (r *patientRepo) SoftDelete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Delete(&entities.Patient{}, "id = ?", id)
	if res.Error != nil {
		r.log.Error("SoftDelete failed", zap.String("patientID", id), zap.Error(res.Error))
	}

	if res.RowsAffected == 0 {
		return ErrNotFound
	}

	r.log.Info("patient soft-deleted", zap.String("patientID", id))
	return nil
}
