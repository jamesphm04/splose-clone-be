package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/jamesphm04/splose-clone-be/internal/models/entities"
	"github.com/jamesphm04/splose-clone-be/internal/repositories"
)

type CreatePatientInput struct {
	Email       string `json:"email" validate:"required,email"`
	FirstName   string `json:"firstName" validate:"required,min=2,max=50"`
	LastName    string `json:"lastName" validate:"required,min=2,max=50"`
	PhoneNumber string `json:"phoneNumber" validate:"required,phoneNumber"`
	DateOfBirth string `json:"dateOfBirth" validate:"required,dateOfBirth"`
	Gender      string `json:"gender" validate:"required,oneof=male female other unknown"`
	FullAddress string `json:"fullAddress" validate:"required,min=2,max=255"`
	UserID      string `json:"userId" validate:"required,uuid"`
}

type UpdatePatientInput struct {
	Email       *string `json:"email" validate:"omitempty,email"`
	FirstName   *string `json:"firstName" validate:"omitempty,min=2,max=50"`
	LastName    *string `json:"lastName" validate:"omitempty,min=2,max=50"`
	PhoneNumber *string `json:"phoneNumber" validate:"omitempty,phoneNumber"`
	DateOfBirth *string `json:"dateOfBirth" validate:"omitempty,dateOfBirth"`
	Gender      *string `json:"gender" validate:"omitempty,oneof=male female other unknown"`
	FullAddress *string `json:"fullAddress" validate:"omitempty,min=2,max=255"`
}

// Service
type PatientService struct {
	repo repositories.PatientRepository
	log  *zap.Logger
}

func NewPatientService(repo repositories.PatientRepository, log *zap.Logger) *PatientService {
	return &PatientService{
		repo: repo,
		log:  log.Named("patient-service"),
	}
}

func (s *PatientService) Create(ctx context.Context, in CreatePatientInput) (*entities.Patient, error) {
	if _, err := s.repo.FindByEmail(ctx, in.Email); err == nil {
		s.log.Warn("patient creation attempt with existing email", zap.String("email", in.Email))
		return nil, ErrEmailTaken
	}

	if _, err := s.repo.FindByPhoneNumber(ctx, in.PhoneNumber); err == nil {
		s.log.Warn("patient creation attempt with existing phone number", zap.String("phoneNumber", in.PhoneNumber))
		return nil, ErrPhoneNumberTaken
	}

	dateOfBirth, err := time.Parse(time.DateOnly, in.DateOfBirth)
	if err != nil {
		s.log.Warn("patient creation attempt with invalid date of birth", zap.String("dateOfBirth", in.DateOfBirth))
		return nil, ErrInvalidDateOfBirth
	}

	gender := entities.Gender(in.Gender)
	if gender != entities.GenderMale && gender != entities.GenderFemale && gender != entities.GenderOther && gender != entities.GenderUnknown {
		s.log.Warn("patient creation attempt with invalid gender", zap.String("gender", in.Gender))
		return nil, ErrInvalidGender
	}

	patient := &entities.Patient{
		Email:       in.Email,
		FirstName:   in.FirstName,
		LastName:    in.LastName,
		PhoneNumber: in.PhoneNumber,
		DateOfBirth: &dateOfBirth,
		Gender:      gender,
		FullAddress: in.FullAddress,
		UserID:      in.UserID,
	}

	if err := s.repo.Create(ctx, patient); err != nil {
		s.log.Error("patient creation failed", zap.Error(err))
		return nil, fmt.Errorf("creating patient: %w", err)
	}

	s.log.Info("patient created", zap.String("email", in.Email))
	return patient, nil
}

func (s *PatientService) GetByID(ctx context.Context, id string) (*entities.Patient, error) {
	patient, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.log.Error("patient retrieval failed", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("retrieving patient: %w", err)
	}
	return patient, nil
}

func (s *PatientService) List(ctx context.Context, offset, limit int) ([]entities.Patient, int64, error) {
	patients, total, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		s.log.Error("patient list failed", zap.Error(err))
		return nil, 0, fmt.Errorf("listing patients: %w", err)
	}
	return patients, total, nil
}

func (s *PatientService) Update(ctx context.Context, id string, in UpdatePatientInput) (*entities.Patient, error) {
	patient, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.log.Error("patient update failed", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("updating patient: %w", err)
	}

	if in.Email != nil {
		patient.Email = *in.Email
	}

	if in.FirstName != nil {
		patient.FirstName = *in.FirstName
	}

	if in.LastName != nil {
		patient.LastName = *in.LastName
	}

	if in.PhoneNumber != nil {
		patient.PhoneNumber = *in.PhoneNumber
	}

	if in.DateOfBirth != nil {
		dateOfBirth, err := time.Parse(time.DateOnly, *in.DateOfBirth)
		if err != nil {
			s.log.Error("patient update failed", zap.String("id", id), zap.Error(err))
			return nil, fmt.Errorf("updating patient: %w", err)
		}
		patient.DateOfBirth = &dateOfBirth
	}

	if in.Gender != nil {
		gender := entities.Gender(*in.Gender)
		if gender != entities.GenderMale && gender != entities.GenderFemale && gender != entities.GenderOther && gender != entities.GenderUnknown {
			s.log.Error("patient update failed", zap.String("id", id), zap.Error(err))
			return nil, fmt.Errorf("updating patient: %w", err)
		}
		patient.Gender = gender
	}

	if in.FullAddress != nil {
		patient.FullAddress = *in.FullAddress
	}

	if err := s.repo.Update(ctx, patient); err != nil {
		s.log.Error("patient update failed", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("updating patient: %w", err)
	}

	s.log.Info("patient updated", zap.String("id", id))
	return patient, nil
}

var (
	ErrPhoneNumberTaken   = errors.New("phone number already taken")
	ErrInvalidGender      = errors.New("invalid gender")
	ErrInvalidDateOfBirth = errors.New("invalid date of birth")
)
