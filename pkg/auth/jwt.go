package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType differentiates access tokens from refresh tokens.
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims extends the standard JWT registered claims with application-specific
// fields. These are embedded in both access and refresh tokens
type Claims struct {
	UserID    string    `json:"id"`
	Role      string    `json:"role"`
	TokenType TokenType `json:"type"`
	jwt.RegisteredClaims
}

// Manager handles token signing and verification
type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewManager creates a Manager with the given HMAC-SHA256 secret and TTLs.
func NewManager(secrete string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		secret:     []byte(secrete),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// GenerateAccessToken mints a short-lived access JWT for the given user
func (m *Manager) GenerateAccessToken(userID, role string) (string, error) {
	return m.generate(userID, role, AccessToken, m.accessTTL)
}

// GenerateRefreshToken mints a long-lived refresh JWT for the given user
func (m *Manager) GenerateRefreshToken(userID, role string) (string, error) {
	return m.generate(userID, role, RefreshToken, m.refreshTTL)
}

func (m *Manager) generate(userID, role string, tt TokenType, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Role:      role,
		TokenType: tt,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signed, nil
}

func (m *Manager) Parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		// Guard: ensure only HS256 is accepted
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return m.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, jwt.ErrTokenExpired
		}

		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// Sentinel errors returned by Parse so callers can switch on them.
var (
	ErrTokenExpired = errors.New("token has expired")
	ErrTokenInvalid = errors.New("token is invalid")
)
