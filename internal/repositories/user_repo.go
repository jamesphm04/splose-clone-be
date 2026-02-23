package repositories

import (
	"context"
	"errors"

	"github.com/jamesphm04/splose-clone-be/internal/models/entities"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	FindByID(ctx context.Context, id string) (*entities.User, error)
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	List(ctx context.Context, offset, limit int) ([]entities.User, int64, error)
	Update(ctx context.Context, user *entities.User) error
	SoftDelete(ctx context.Context, id string) error
}

type userRepo struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewUserRepository returns a GORM-backed UserRepository.
func NewUserRepository(db *gorm.DB, log *zap.Logger) UserRepository {
	return &userRepo{
		db:  db,
		log: log.Named("user-repository"),
	}
}

func (r *userRepo) Create(ctx context.Context, user *entities.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		r.log.Error("failed to create user", zap.String("email", user.Email), zap.Error(err))
		return err
	}

	r.log.Info("user created", zap.String("email", user.Email))
	return nil
}

func (r *userRepo) FindByID(ctx context.Context, id string) (*entities.User, error) {
	var u entities.User
	err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		r.log.Error("FindByID failed", zap.String("id", id), zap.Error(err))
	}

	return &u, nil
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	var u entities.User
	err := r.db.WithContext(ctx).First(&u, "email = ?", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		r.log.Error("FindByEmail failed", zap.String("email", email), zap.Error(err))
	}
	return &u, nil
}

func (r *userRepo) List(ctx context.Context, offset, limit int) ([]entities.User, int64, error) {
	var users []entities.User
	var total int64

	// count total
	if err := r.db.WithContext(ctx).Model(&entities.User{}).Count(&total).Error; err != nil {
		r.log.Error("List count failed", zap.Error(err))
		return nil, 0, err
	}

	// list
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		r.log.Error("List query failed", zap.Error(err))
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepo) Update(ctx context.Context, user *entities.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		r.log.Error("Update failed", zap.String("userID", user.ID), zap.Error(err))
		return err
	}

	return nil
}

func (r *userRepo) SoftDelete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Delete(&entities.User{}, "id = ?", id)
	if res.Error != nil {
		r.log.Error("SoftDelete failed", zap.String("userID", id), zap.Error(res.Error))
	}

	if res.RowsAffected == 0 {
		return ErrNotFound
	}

	r.log.Info("user soft-deleted", zap.String("userID", id))
	return nil
}

// Share between repos
var (
	ErrNotFound     = errors.New("record not found")
	ErrDuplicateKey = errors.New("duplicate key")
)
