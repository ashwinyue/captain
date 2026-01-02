package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/pkg/jwt"
	"github.com/tgo/captain/apiserver/internal/pkg/wukongim"
)

type AuthService struct {
	db             *gorm.DB
	jwtManager     *jwt.Manager
	wukongimClient *wukongim.Client
}

func NewAuthService(db *gorm.DB, jwtManager *jwt.Manager, wukongimClient *wukongim.Client) *AuthService {
	return &AuthService{db: db, jwtManager: jwtManager, wukongimClient: wukongimClient}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int          `json:"expires_in"`
	Staff        *model.Staff `json:"staff"`
}

func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*TokenResponse, error) {
	var staff model.Staff
	if err := s.db.WithContext(ctx).Where("username = ? AND is_active = true AND deleted_at IS NULL", req.Username).First(&staff).Error; err != nil {
		return nil, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(staff.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	accessToken, err := s.jwtManager.GenerateAccessToken(staff.ID, staff.Username, staff.Role, staff.ProjectID)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.jwtManager.GenerateRefreshToken(staff.ID, staff.Username)
	if err != nil {
		return nil, err
	}

	// Update staff status to online
	staff.Status = "online"
	s.db.WithContext(ctx).Model(&staff).Update("status", "online")

	// Register user with WuKongIM for WebSocket authentication
	if s.wukongimClient != nil {
		staffUID := fmt.Sprintf("%s-staff", staff.ID.String())
		_, err := s.wukongimClient.GetToken(ctx, &wukongim.GetTokenRequest{
			UID:         staffUID,
			Token:       accessToken, // Use JWT token as WuKongIM token
			DeviceFlag:  1,           // 1 = web
			DeviceLevel: 1,           // 1 = primary
		})
		if err != nil {
			log.Printf("Failed to register staff %s with WuKongIM: %v", staffUID, err)
			// Don't fail login if WuKongIM registration fails
		} else {
			log.Printf("Successfully registered staff %s with WuKongIM", staffUID)
		}
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
		ExpiresIn:    3600,
		Staff:        &staff,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}
	var staff model.Staff
	if err := s.db.WithContext(ctx).Where("id = ? AND is_active = true AND deleted_at IS NULL", claims.UserID).First(&staff).Error; err != nil {
		return nil, errors.New("user not found")
	}
	accessToken, err := s.jwtManager.GenerateAccessToken(staff.ID, staff.Username, staff.Role, staff.ProjectID)
	if err != nil {
		return nil, err
	}
	newRefreshToken, err := s.jwtManager.GenerateRefreshToken(staff.ID, staff.Username)
	if err != nil {
		return nil, err
	}
	return &TokenResponse{AccessToken: accessToken, RefreshToken: newRefreshToken, TokenType: "bearer"}, nil
}

func (s *AuthService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*model.Staff, error) {
	var staff model.Staff
	if err := s.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", userID).First(&staff).Error; err != nil {
		return nil, err
	}
	return &staff, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
