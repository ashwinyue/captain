package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Manager struct {
	secret                 []byte
	accessTokenExpireMin   int
	refreshTokenExpireDays int
}

type Claims struct {
	jwt.RegisteredClaims
	UserID    uuid.UUID  `json:"user_id"`
	Username  string     `json:"username"`
	Role      string     `json:"role"`
	ProjectID *uuid.UUID `json:"project_id,omitempty"`
}

func NewManager(secret string, accessExpMin, refreshExpDays int) *Manager {
	return &Manager{
		secret:                 []byte(secret),
		accessTokenExpireMin:   accessExpMin,
		refreshTokenExpireDays: refreshExpDays,
	}
}

func (m *Manager) GenerateAccessToken(userID uuid.UUID, username, role string, projectID *uuid.UUID) (string, error) {
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(m.accessTokenExpireMin) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   username,
		},
		UserID:    userID,
		Username:  username,
		Role:      role,
		ProjectID: projectID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) GenerateRefreshToken(userID uuid.UUID, username string) (string, error) {
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(m.refreshTokenExpireDays) * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   username,
		},
		UserID:   userID,
		Username: username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
