package services

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/jamesphm04/splose-clone-be/internal/models/entities"
	"github.com/jamesphm04/splose-clone-be/internal/repositories"
	"github.com/jamesphm04/splose-clone-be/pkg/auth"
)

type RegisterInput struct {
	Email    string `json:"email"    validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginInput struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type UpdateUserInput struct {
	Username string `json:"username" validate:"omitempty,min=3,max=50"`
	Email    string `json:"email"    validate:"omitempty,email"`
}

// Service
type UserService struct {
	repo       repositories.UserRepository
	jwtManager *auth.Manager
	bcryptCost int
	log        *zap.Logger
}

func NewUserService(
	repo repositories.UserRepository,
	jwtManager *auth.Manager,
	bcryptCost int,
	log *zap.Logger,
) *UserService {
	return &UserService{
		repo:       repo,
		jwtManager: jwtManager,
		bcryptCost: bcryptCost,
		log:        log.Named("user_service"),
	}
}

// Register creates a new user after hashing the password
func (s *UserService) Register(ctx context.Context, in RegisterInput) (*entities.User, error) {
	if _, err := s.repo.FindByEmail(ctx, in.Email); err == nil {
		s.log.Warn("registration attempt with existing email", zap.String("email", in.Email))
		return nil, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), s.bcryptCost)
	if err != nil {
		s.log.Error("bcrypt failed during registration", zap.Error(err))
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	// Build user
	user := &entities.User{
		Email:        in.Email,
		Username:     in.Username,
		PasswordHash: string(hash),
		Role:         "user",
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	s.log.Info("user registered", zap.String("userID", user.ID), zap.String("email", user.Email))
	return user, nil
}

// Login authenticates credentials and returns a JWT token pair
func (s *UserService) Login(ctx context.Context, in LoginInput) (*entities.User, *TokenPair, error) {
	user, err := s.repo.FindByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			// Not and error, just a fail attempt
			s.log.Debug("login failed: email not found", zap.String("email", in.Email))
			return nil, nil, err
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
		// Not and error, just a fail attempt
		s.log.Debug("login failed: wrong password", zap.String("userID", user.ID))
		return nil, nil, ErrInvalidCredentials
	}

	pair, err := s.issueTokenPair(user.ID, user.Role)
	if err != nil {
		return nil, nil, err
	}

	s.log.Info("user logged in", zap.String("userID", user.ID))
	return user, pair, nil
}

// RefreshTokens validates a refresh token and issues a new pair
func (s *UserService) RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.jwtManager.Parse(refreshToken)
	if err != nil {
		s.log.Debug("token refresh failed: invalid token", zap.Error(err))
		return nil, ErrInvalidRefreshToken
	}

	if claims.TokenType != auth.RefreshToken {
		return nil, ErrInvalidRefreshToken
	}

	user, err := s.repo.FindByEmail(ctx, claims.UserID)
	if err != nil {
		s.log.Warn("token refresh failed: user not found", zap.String("userID", claims.UserID))
		return nil, ErrInvalidCredentials
	}

	// Create a new pair of token
	pair, err := s.issueTokenPair(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	s.log.Info("tokens refreshed", zap.String("userID", user.ID))
	return pair, nil
}

func (s *UserService) GetByID(ctx context.Context, id string) (*entities.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *UserService) List(ctx context.Context, offset, limit int) ([]entities.User, int64, error) {
	return s.repo.List(ctx, offset, limit)
}

func (s *UserService) Update(ctx context.Context, id string, in UpdateUserInput) (*entities.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding user by id: %w", err)
	}

	if in.Username != "" {
		user.Username = in.Username
	}

	if in.Email != "" {
		user.Email = in.Email
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	s.log.Info("user updated", zap.String("userID", id))
	return user, nil
}

func (s *UserService) SoftDelete(ctx context.Context, id string) error {
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return err
	}
	s.log.Info("user deleted", zap.String("userID", id))
	return nil
}

// issueTokenPair generate a pair of tokens for an authenticated user
func (s *UserService) issueTokenPair(userID, role string) (*TokenPair, error) {
	access, err := s.jwtManager.GenerateAccessToken(userID, role)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}
	refresh, err := s.jwtManager.GenerateRefreshToken(userID, role)
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}
	return &TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

var (
	ErrEmailTaken          = errors.New("email already taken")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)
